package withdraw

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type FPWithdrawer struct {
	Ctx           context.Context
	L1Client      *ethclient.Client
	L2Client      *rpc.Client
	L2TxHash      common.Hash
	Portal        *bindingspreview.OptimismPortal2
	Factory       *bindings.DisputeGameFactory
	Opts          *bind.TransactOpts
	GasMultiplier float64 // Multiplier for estimated gas (default 1.0)
}

func (w *FPWithdrawer) CheckIfProvable() error {
	l2WithdrawalBlock, err := txBlock(w.Ctx, w.L2Client, w.L2TxHash)
	if err != nil {
		return fmt.Errorf("error querying withdrawal tx block: %w", err)
	}

	latestGame, err := withdrawals.FindLatestGame(w.Ctx, &w.Factory.DisputeGameFactoryCaller, &w.Portal.OptimismPortal2Caller)
	if err != nil {
		return fmt.Errorf("failed to find latest game: %w", err)
	}
	l2BlockNumber := new(big.Int).SetBytes(latestGame.ExtraData[0:32])

	if l2BlockNumber.Uint64() < l2WithdrawalBlock.Uint64() {
		return fmt.Errorf("the latest L2 block proposed in the DisputeGameFactory is %d and is not past L2 block %d that includes the withdrawal - the withdrawal cannot be proven yet",
			l2BlockNumber.Uint64(), l2WithdrawalBlock.Uint64())
	}
	return nil
}

func (w *FPWithdrawer) getWithdrawalHash() (common.Hash, error) {
	l2 := ethclient.NewClient(w.L2Client)
	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return common.HexToHash(""), err
	}

	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		return common.HexToHash(""), err
	}

	hash, err := withdrawals.WithdrawalHash(ev)
	if err != nil {
		return common.HexToHash(""), err
	}

	return hash, nil
}

func (w *FPWithdrawer) GetProvenWithdrawalTime() (uint64, error) {
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return 0, err
	}

	// the proven withdrawal structure now contains an additional mapping, as withdrawal proofs are now stored per submitter address
	provenWithdrawal, err := w.Portal.ProvenWithdrawals(&bind.CallOpts{}, hash, w.Opts.From)
	if err != nil {
		return 0, err
	}

	return provenWithdrawal.Timestamp, nil
}

func (w *FPWithdrawer) ProveWithdrawal() error {
	l2 := ethclient.NewClient(w.L2Client)
	l2g := gethclient.New(w.L2Client)

	params, err := withdrawals.ProveWithdrawalParametersFaultProofs(w.Ctx, l2g, l2, l2, w.L2TxHash, &w.Factory.DisputeGameFactoryCaller, &w.Portal.OptimismPortal2Caller)
	if err != nil {
		return err
	}

	withdrawalTx := bindingspreview.TypesWithdrawalTransaction{
		Nonce:    params.Nonce,
		Sender:   params.Sender,
		Target:   params.Target,
		Value:    params.Value,
		GasLimit: params.GasLimit,
		Data:     params.Data,
	}
	outputRootProof := bindingspreview.TypesOutputRootProof{
		Version:                  params.OutputRootProof.Version,
		StateRoot:                params.OutputRootProof.StateRoot,
		MessagePasserStorageRoot: params.OutputRootProof.MessagePasserStorageRoot,
		LatestBlockhash:          params.OutputRootProof.LatestBlockhash,
	}

	// Apply gas multiplier if set and no explicit gas limit
	if w.GasMultiplier > 1.0 && w.Opts.GasLimit == 0 {
		// Estimate gas using NoSend
		estimateOpts := *w.Opts
		estimateOpts.NoSend = true
		estimateTx, err := w.Portal.ProveWithdrawalTransaction(
			&estimateOpts,
			withdrawalTx,
			params.L2OutputIndex,
			outputRootProof,
			params.WithdrawalProof,
		)
		if err != nil {
			return fmt.Errorf("failed to estimate gas: %w", err)
		}
		adjustedGas := uint64(float64(estimateTx.Gas()) * w.GasMultiplier)
		w.Opts.GasLimit = adjustedGas
		log.Info("Adjusted gas estimate", "original", estimateTx.Gas(), "multiplier", w.GasMultiplier, "adjusted", adjustedGas)
	}

	// create the proof
	tx, err := w.Portal.ProveWithdrawalTransaction(
		w.Opts,
		withdrawalTx,
		params.L2OutputIndex, // this is overloaded and is the DisputeGame index in this context
		outputRootProof,
		params.WithdrawalProof,
	)
	if err != nil {
		return err
	}

	fmt.Printf("Proved withdrawal for %s: %s\n", w.L2TxHash.String(), tx.Hash().String())

	// Wait 5 mins max for confirmation
	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()
	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash())
}

func (w *FPWithdrawer) IsProofFinalized() (bool, error) {
	return w.Portal.FinalizedWithdrawals(&bind.CallOpts{}, w.L2TxHash)
}

func (w *FPWithdrawer) FinalizeWithdrawal() error {
	// get the withdrawal hash
	hash, err := w.getWithdrawalHash()
	if err != nil {
		return err
	}

	// check if the withdrawal can be finalized using the calculated withdrawal hash
	err = w.Portal.CheckWithdrawal(&bind.CallOpts{}, hash, w.Opts.From)
	if err != nil {
		return err
	}

	// get the WithdrawalTransaction info needed to finalize the withdrawal
	l2 := ethclient.NewClient(w.L2Client)

	// Transaction receipt
	receipt, err := l2.TransactionReceipt(w.Ctx, w.L2TxHash)
	if err != nil {
		return err
	}
	// Parse the receipt
	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		return err
	}

	withdrawalTx := bindingspreview.TypesWithdrawalTransaction{
		Nonce:    ev.Nonce,
		Sender:   ev.Sender,
		Target:   ev.Target,
		Value:    ev.Value,
		GasLimit: ev.GasLimit,
		Data:     ev.Data,
	}

	// Apply gas multiplier if set and no explicit gas limit
	if w.GasMultiplier > 1.0 && w.Opts.GasLimit == 0 {
		// Estimate gas using NoSend
		estimateOpts := *w.Opts
		estimateOpts.NoSend = true
		estimateTx, err := w.Portal.FinalizeWithdrawalTransaction(&estimateOpts, withdrawalTx)
		if err != nil {
			return fmt.Errorf("failed to estimate gas: %w", err)
		}
		adjustedGas := uint64(float64(estimateTx.Gas()) * w.GasMultiplier)
		w.Opts.GasLimit = adjustedGas
		log.Info("Adjusted gas estimate", "original", estimateTx.Gas(), "multiplier", w.GasMultiplier, "adjusted", adjustedGas)
	}

	// finalize the withdrawal
	tx, err := w.Portal.FinalizeWithdrawalTransaction(w.Opts, withdrawalTx)
	if err != nil {
		return err
	}

	fmt.Printf("Completed withdrawal for %s: %s\n", w.L2TxHash.String(), tx.Hash().String())

	// Wait 5 mins max for confirmation
	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()
	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash())
}
