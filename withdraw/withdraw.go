package withdraw

import (
	"context"
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
	GasMultiplier float64      // Multiplier for estimated gas (default 1.0)
	UserGasLimit  uint64       // Original user-specified gas limit (0 means auto-estimate)
	DryRun        bool         // Simulate transactions without submitting
	Retry         *RetryConfig // Bind Portal and Oracle with NewRetryingBackend using this config.
}

func (w *Withdrawer) CheckIfProvable() error {
	// check to make sure it is possible to prove the provided withdrawal
	submissionInterval, err := w.Oracle.SUBMISSIONINTERVAL(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return fmt.Errorf("error querying output proposal submission interval: %w", err)
	}

	l2BlockTime, err := w.Oracle.L2BLOCKTIME(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return fmt.Errorf("error querying output proposal L2 block time: %w", err)
	}

	l2OutputBlock, err := w.Oracle.LatestBlockNumber(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return fmt.Errorf("error querying latest proposed block: %w", err)
	}

	l2WithdrawalBlock, err := txBlock(w.Ctx, w.L2Client, w.L2TxHash, w.Retry)
	if err != nil {
		return fmt.Errorf("error querying withdrawal tx block: %w", err)
	}

	if l2OutputBlock.Uint64() < l2WithdrawalBlock.Uint64() {
		return fmt.Errorf("the latest L2 output is %d and is not past L2 block %d that includes the withdrawal, no withdrawal can be proved yet - please wait for the next proposal submission, which happens every %v",
			l2OutputBlock.Uint64(), l2WithdrawalBlock.Uint64(), time.Duration(submissionInterval.Int64()*l2BlockTime.Int64())*time.Second)
	}
	return nil
}

func (w *Withdrawer) getWithdrawalHash() (common.Hash, error) {
	l2 := retryingL2{Client: ethclient.NewClient(w.L2Client), config: w.Retry}
	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return common.Hash{}, err
	}

	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		return common.Hash{}, err
	}

	hash, err := withdrawals.WithdrawalHash(ev)
	if err != nil {
		return common.Hash{}, err
	}

	return hash, nil
}

func (w *Withdrawer) GetProvenWithdrawalTime() (uint64, error) {
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return 0, err
	}

	provenWithdrawal, err := w.Portal.ProvenWithdrawals(&bind.CallOpts{Context: w.Ctx}, hash)
	if err != nil {
		return 0, err
	}

	return provenWithdrawal.Timestamp.Uint64(), nil
}

func (w *Withdrawer) ProveWithdrawal() error {
	l2 := retryingL2{Client: ethclient.NewClient(w.L2Client), proof: gethclient.New(w.L2Client), config: w.Retry}

	l2OutputBlock, err := w.Oracle.LatestBlockNumber(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return err
	}

	// We generate a proof for the latest L2 output, which shouldn't require archive-node data if it's recent enough.
	header, err := l2.HeaderByNumber(w.Ctx, l2OutputBlock)
	if err != nil {
		return err
	}
	params, err := withdrawals.ProveWithdrawalParameters(w.Ctx, l2, l2, w.L2TxHash, header, &w.Oracle.L2OutputOracleCaller)
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

	// Prepare gas options with multiplier if configured
	simulatedTx, err := prepareGasOpts(w.Opts, w.UserGasLimit, w.GasMultiplier, w.DryRun, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return w.Portal.ProveWithdrawalTransaction(
			opts,
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
		printDryRun("ProveWithdrawal", simulatedTx, w.Opts.From, w.Opts.GasLimit)
		return nil
	}

	noSendOpts := *w.Opts
	noSendOpts.NoSend = true
	tx, err := w.Portal.ProveWithdrawalTransaction(
		&noSendOpts,
		withdrawalTx,
		params.L2OutputIndex,
		params.OutputRootProof,
		params.WithdrawalProof,
	)
	if err != nil {
		return err
	}
	if w.Opts.NoSend {
		log.Info("Prepared withdrawal transaction without submission", "l1TxHash", tx.Hash())
		return nil
	}
	if err := sendTransaction(w.Ctx, w.Opts.Context, w.L1Client, w.Retry, "ProveWithdrawal.send", tx); err != nil {
		return err
	}

	log.Info("Proved withdrawal", "l2TxHash", w.L2TxHash, "l1TxHash", tx.Hash())

	// Wait 5 mins max for confirmation
	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()
	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash(), "proof", w.Retry)
}

func (w *Withdrawer) IsProofFinalized() (bool, error) {
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return false, err
	}
	return w.Portal.FinalizedWithdrawals(&bind.CallOpts{Context: w.Ctx}, hash)
}

func (w *Withdrawer) FinalizeWithdrawal() error {
	l2 := retryingL2{Client: ethclient.NewClient(w.L2Client), proof: gethclient.New(w.L2Client), config: w.Retry}

	// Figure out when our withdrawal was included
	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return fmt.Errorf("failed to fetch receipt for L2 withdrawal transaction %s: %w", w.L2TxHash, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return fmt.Errorf("L2 withdrawal transaction failed: tx %s, receipt status %d", w.L2TxHash, receipt.Status)
	}

	l2WithdrawalBlock, err := l2.HeaderByNumber(w.Ctx, receipt.BlockNumber)
	if err != nil {
		return fmt.Errorf("error getting header by number for block %s: %w", receipt.BlockNumber, err)
	}

	// Figure out what the Output oracle on L1 has seen so far
	l2OutputBlockNr, err := w.Oracle.LatestBlockNumber(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return err
	}

	l2OutputBlock, err := l2.HeaderByNumber(w.Ctx, l2OutputBlockNr)
	if err != nil {
		return fmt.Errorf("error getting header by number for latest block %s: %w", l2OutputBlockNr, err)
	}

	// Check if the L2 output is even old enough to include the withdrawal
	if l2OutputBlock.Number.Uint64() < l2WithdrawalBlock.Number.Uint64() {
		return fmt.Errorf("the latest L2 output is %d and is not past L2 block %d that includes the withdrawal yet, no withdrawal can be completed yet", l2OutputBlock.Number.Uint64(), l2WithdrawalBlock.Number.Uint64())
	}

	l1Head, err := NewRetryingBackend(w.Ctx, w.L1Client, w.Retry).HeaderByNumber(w.Ctx, nil)
	if err != nil {
		return err
	}

	// Check if the withdrawal may be completed yet
	finalizationPeriod, err := w.Oracle.FINALIZATIONPERIODSECONDS(&bind.CallOpts{Context: w.Ctx})
	if err != nil {
		return err
	}

	if l2WithdrawalBlock.Time+finalizationPeriod.Uint64() >= l1Head.Time {
		return fmt.Errorf("withdrawal tx %s was included in L2 block %d (time %d) but L1 only knows of L2 proposal %d (time %d) at head %d (time %d) which has not reached output confirmation yet (period is %d)",
			w.L2TxHash, l2WithdrawalBlock.Number.Uint64(), l2WithdrawalBlock.Time, l2OutputBlock.Number.Uint64(), l2OutputBlock.Time, l1Head.Number.Uint64(), l1Head.Time, finalizationPeriod.Uint64())
	}

	// We generate a proof for the latest L2 output, which shouldn't require archive-node data if it's recent enough.
	// Note that for the `FinalizeWithdrawalTransaction` function, this proof isn't needed. We simply use some of the
	// params for the `WithdrawalTransaction` type generated in the bindings.
	header, err := l2.HeaderByNumber(w.Ctx, l2OutputBlockNr)
	if err != nil {
		return err
	}

	params, err := withdrawals.ProveWithdrawalParameters(w.Ctx, l2, l2, w.L2TxHash, header, &w.Oracle.L2OutputOracleCaller)
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

	// Prepare gas options with multiplier if configured
	simulatedTx, err := prepareGasOpts(w.Opts, w.UserGasLimit, w.GasMultiplier, w.DryRun, func(opts *bind.TransactOpts) (*types.Transaction, error) {
		return w.Portal.FinalizeWithdrawalTransaction(opts, withdrawalTx)
	})
	if err != nil {
		return err
	}

	if w.DryRun {
		printDryRun("FinalizeWithdrawal", simulatedTx, w.Opts.From, w.Opts.GasLimit)
		return nil
	}

	noSendOpts := *w.Opts
	noSendOpts.NoSend = true
	tx, err := w.Portal.FinalizeWithdrawalTransaction(&noSendOpts, withdrawalTx)
	if err != nil {
		return err
	}

	if w.Opts.NoSend {
		log.Info("Prepared withdrawal transaction without submission", "l1TxHash", tx.Hash())
		return nil
	}
	if err := sendTransaction(w.Ctx, w.Opts.Context, w.L1Client, w.Retry, "FinalizeWithdrawal.send", tx); err != nil {
		return err
	}

	log.Info("Completed withdrawal", "l2TxHash", w.L2TxHash, "l1TxHash", tx.Hash())

	// Wait 5 mins max for confirmation
	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()
	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash(), "finalization", w.Retry)
}
