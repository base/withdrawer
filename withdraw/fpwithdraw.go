package withdraw

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

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
	Ctx      context.Context
	L1Client *ethclient.Client
	L2Client *rpc.Client
	L2TxHash common.Hash
	Portal   *bindingspreview.OptimismPortal2
	Factory  *bindings.DisputeGameFactory
	Opts     *bind.TransactOpts
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

	params, err := ProveWithdrawalParametersFaultProofs(
		w.Ctx,
		l2g,
		l2,
		l2,
		w.L2TxHash,
		&w.Factory.DisputeGameFactoryCaller,
		&w.Portal.OptimismPortal2Caller,
	)
	if err != nil {
		return err
	}

	// create the proof
	tx, err := w.Portal.ProveWithdrawalTransaction(
		w.Opts,
		bindingspreview.TypesWithdrawalTransaction{
			Nonce:    params.Nonce,
			Sender:   params.Sender,
			Target:   params.Target,
			Value:    params.Value,
			GasLimit: params.GasLimit,
			Data:     params.Data,
		},
		params.L2OutputIndex, // this is overloaded and is the DisputeGame index in this context
		bindingspreview.TypesOutputRootProof{
			Version:                  params.OutputRootProof.Version,
			StateRoot:                params.OutputRootProof.StateRoot,
			MessagePasserStorageRoot: params.OutputRootProof.MessagePasserStorageRoot,
			LatestBlockhash:          params.OutputRootProof.LatestBlockhash,
		},
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

	// finalize the withdrawal
	tx, err := w.Portal.FinalizeWithdrawalTransaction(
		w.Opts,
		bindingspreview.TypesWithdrawalTransaction{
			Nonce:    ev.Nonce,
			Sender:   ev.Sender,
			Target:   ev.Target,
			Value:    ev.Value,
			GasLimit: ev.GasLimit,
			Data:     ev.Data,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("Completed withdrawal for %s: %s\n", w.L2TxHash.String(), tx.Hash().String())

	// Wait 5 mins max for confirmation
	ctxWithTimeout, cancel := context.WithTimeout(w.Ctx, 5*time.Minute)
	defer cancel()
	return waitForConfirmation(ctxWithTimeout, w.L1Client, tx.Hash())
}

// ProveWithdrawalParametersFaultProofs calls ProveWithdrawalParametersForBlock with the most recent L2 output after the latest game.
func ProveWithdrawalParametersFaultProofs(ctx context.Context, proofCl withdrawals.ProofClient, l2ReceiptCl withdrawals.ReceiptClient, l2BlockCl withdrawals.HeaderClient, txHash common.Hash, disputeGameFactoryContract *bindings.DisputeGameFactoryCaller, optimismPortal2Contract *bindingspreview.OptimismPortal2Caller) (withdrawals.ProvenWithdrawalParameters, error) {
	latestGame, err := FindEarliestGame(ctx, l2ReceiptCl, txHash, disputeGameFactoryContract, optimismPortal2Contract)
	if err != nil {
		return withdrawals.ProvenWithdrawalParameters{}, fmt.Errorf("failed to find game: %w", err)
	}

	l2BlockNumber := new(big.Int).SetBytes(latestGame.ExtraData[0:32])
	l2Header, err := l2BlockCl.HeaderByNumber(ctx, l2BlockNumber)
	if err != nil {
		return withdrawals.ProvenWithdrawalParameters{}, fmt.Errorf("failed to get l2Block: %w", err)
	}

	l2OutputIndex := latestGame.Index
	return withdrawals.ProveWithdrawalParametersForBlock(ctx, proofCl, l2ReceiptCl, txHash, l2Header, l2OutputIndex)
}

// FindEarliestGame finds the earliest game in the DisputeGameFactory contract that is after the given transaction receipt.
// Note that this does not support checking for invalid games (e.g. games that were successfully challenged).
func FindEarliestGame(ctx context.Context, l2ReceiptCl withdrawals.ReceiptClient, txHash common.Hash, disputeGameFactoryContract *bindings.DisputeGameFactoryCaller, optimismPortal2Contract *bindingspreview.OptimismPortal2Caller) (*bindings.IDisputeGameFactoryGameSearchResult, error) {
	receipt, err := l2ReceiptCl.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	respectedGameType, err := optimismPortal2Contract.RespectedGameType(&bind.CallOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get respected game type: %w", err)
	}

	gameCount, err := disputeGameFactoryContract.GameCount(&bind.CallOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get game count: %w", err)
	}
	if gameCount.Cmp(common.Big0) == 0 {
		return nil, errors.New("no games")
	}

	lo := new(big.Int)
	hi := new(big.Int).Sub(gameCount, common.Big1)

	for lo.Cmp(hi) < 0 {
		mid := new(big.Int).Add(lo, hi)
		mid.Div(mid, common.Big2)
		latestGames, err := disputeGameFactoryContract.FindLatestGames(&bind.CallOpts{}, respectedGameType, mid, common.Big1)
		if err != nil {
			return nil, err
		}
		l2BlockNumber := new(big.Int)
		if len(latestGames) > 0 {
			l2BlockNumber = new(big.Int).SetBytes(latestGames[0].ExtraData[0:32])
		}
		if l2BlockNumber.Cmp(receipt.BlockNumber) < 0 {
			lo = mid.Add(mid, common.Big1)
		} else {
			hi = mid
		}
	}

	latestGames, err := disputeGameFactoryContract.FindLatestGames(&bind.CallOpts{}, respectedGameType, lo, common.Big1)
	if err != nil {
		return nil, err
	}
	if len(latestGames) == 0 {
		return nil, errors.New("no games found")
	}
	latestGame := latestGames[0]

	return &latestGame, nil
}
