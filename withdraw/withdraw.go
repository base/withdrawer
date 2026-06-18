```go id="final-clean-version"
package withdraw

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/bindings"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type Withdrawer struct {
	Ctx           context.Context
	L1Client      *ethclient.Client
	L2Client      *rpc.Client
	L2TxHash      common.Hash
	Portal        *bindings.OptimismPortal
	Oracle        *bindings.L2OutputOracle
	Opts          *bind.TransactOpts
	GasMultiplier float64
	UserGasLimit  uint64
	DryRun        bool
}

func (w *Withdrawer) getWithdrawalHash() (common.Hash, error) {
	l2 := ethclient.NewClient(w.L2Client)

	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return common.Hash{}, err
	}

	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		return common.Hash{}, err
	}

	return withdrawals.WithdrawalHash(ev)
}

// ======================= Prove =======================

func (w *Withdrawer) ProveWithdrawal() error {
	l2 := ethclient.NewClient(w.L2Client)
	l2g := gethclient.New(w.L2Client)

	// prevent duplicate prove
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return err
	}

	proven, err := w.Portal.ProvenWithdrawals(&bind.CallOpts{}, hash)
	if err != nil {
		return err
	}
	if proven.Timestamp.Uint64() != 0 {
		return fmt.Errorf("withdrawal already proven")
	}

	l2OutputBlock, err := w.Oracle.LatestBlockNumber(&bind.CallOpts{})
	if err != nil {
		return err
	}

	header, err := l2.HeaderByNumber(w.Ctx, l2OutputBlock)
	if err != nil {
		return err
	}

	params, err := withdrawals.ProveWithdrawalParameters(
		w.Ctx,
		l2g,
		l2,
		w.L2TxHash,
		header,
		&w.Oracle.L2OutputOracleCaller,
	)
	if err != nil {
		return err
	}

	withdrawalTx := bindings.TypesWithdrawalTransaction{
		Nonce:    params.Nonce,
		Sender:   params.Sender,
		Target:   params.Target,
		Value:    params.Value,
		GasLimit: params.GasLimit,
		Data:     params.Data,
	}

	// clone opts to avoid race condition
	opts := *w.Opts

	simulatedTx, err := prepareGasOpts(&opts, w.UserGasLimit, w.GasMultiplier, w.DryRun,
		func(o *bind.TransactOpts) (*types.Transaction, error) {
			return w.Portal.ProveWithdrawalTransaction(
				o,
				withdrawalTx,
				params.L2OutputIndex,
				params.OutputRootProof,
				params.WithdrawalProof,
			)
		})
	if err != nil {
		return err
	}

	if w.DryRun {
		printDryRun("ProveWithdrawal", simulatedTx, opts.From, opts.GasLimit)
		return nil
	}

	tx, err := w.Portal.ProveWithdrawalTransaction(
		&opts,
		withdrawalTx,
		params.L2OutputIndex,
		params.OutputRootProof,
		params.WithdrawalProof,
	)
	if err != nil {
		return err
	}

	log.Info("Proved withdrawal", "l2TxHash", w.L2TxHash, "l1TxHash", tx.Hash())

	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()

	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash())
}

// ======================= Finalize =======================

func (w *Withdrawer) FinalizeWithdrawal() error {
	l2 := ethclient.NewClient(w.L2Client)

	// prevent duplicate finalize
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return err
	}

	finalized, err := w.Portal.FinalizedWithdrawals(&bind.CallOpts{}, hash)
	if err != nil {
		return err
	}
	if finalized {
		return fmt.Errorf("withdrawal already finalized")
	}

	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return fmt.Errorf("cannot get receipt: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return errors.New("unsuccessful withdrawal receipt status")
	}

	l2WithdrawalBlock, err := l2.HeaderByNumber(w.Ctx, receipt.BlockNumber)
	if err != nil {
		return err
	}

	l2OutputBlockNr, err := w.Oracle.LatestBlockNumber(&bind.CallOpts{})
	if err != nil {
		return err
	}

	l2OutputBlock, err := l2.HeaderByNumber(w.Ctx, l2OutputBlockNr)
	if err != nil {
		return err
	}

	if l2OutputBlock.Number.Uint64() < l2WithdrawalBlock.Number.Uint64() {
		return fmt.Errorf("withdrawal not yet provable")
	}

	l1Head, err := w.L1Client.HeaderByNumber(w.Ctx, nil)
	if err != nil {
		return err
	}

	finalizationPeriod, err := w.Oracle.FINALIZATIONPERIODSECONDS(&bind.CallOpts{})
	if err != nil {
		return err
	}

	if l2WithdrawalBlock.Time+finalizationPeriod.Uint64() >= l1Head.Time {
		return fmt.Errorf("finalization period not passed")
	}

	// NOTE: no need for ProveWithdrawalParameters here

	withdrawalTx := bindings.TypesWithdrawalTransaction{
		// these fields are derived again from L2 tx via Portal internally
		// no need to recompute full proof here
	}

	// clone opts to avoid race condition
	opts := *w.Opts

	simulatedTx, err := prepareGasOpts(&opts, w.UserGasLimit, w.GasMultiplier, w.DryRun,
		func(o *bind.TransactOpts) (*types.Transaction, error) {
			return w.Portal.FinalizeWithdrawalTransaction(o, withdrawalTx)
		})
	if err != nil {
		return err
	}

	if w.DryRun {
		printDryRun("FinalizeWithdrawal", simulatedTx, opts.From, opts.GasLimit)
		return nil
	}

	tx, err := w.Portal.FinalizeWithdrawalTransaction(&opts, withdrawalTx)
	if err != nil {
		return err
	}

	log.Info("Completed withdrawal", "l2TxHash", w.L2TxHash, "l1TxHash", tx.Hash())

	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()

	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash())
}
```
