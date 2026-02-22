package withdraw

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type WithdrawHelper interface {
	CheckIfProvable() error
	GetProvenWithdrawalTime() (uint64, error)
	ProveWithdrawal() error
	IsProofFinalized() (bool, error)
	FinalizeWithdrawal() error
}

func txBlock(ctx context.Context, l2c *rpc.Client, l2TxHash common.Hash) (*big.Int, error) {
	l2 := ethclient.NewClient(l2c)
	// Figure out when our withdrawal was included
	receipt, err := l2.TransactionReceipt(ctx, l2TxHash)
	if err != nil {
		return nil, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("unsuccessful withdrawal receipt status")
	}
	return receipt.BlockNumber, nil
}

func waitForConfirmation(ctx context.Context, client *ethclient.Client, tx common.Hash) error {
	for {
		receipt, err := client.TransactionReceipt(ctx, tx)
		if err == ethereum.NotFound {
			log.Info("Waiting for tx confirmation", "txHash", tx.String())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		} else if err != nil {
			return err
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return errors.New("unsuccessful withdrawal receipt status")
		} else {
			break
		}
	}
	log.Info("Transaction confirmed", "txHash", tx.String())
	return nil
}

// prepareGasOpts resets the gas limit, applies gas multiplier if needed, and
// optionally simulates the transaction for dry-run mode. The simulateFn should
// perform a NoSend transaction and return the resulting *types.Transaction.
// Returns the simulated tx when a simulation was performed, or nil otherwise.
func prepareGasOpts(opts *bind.TransactOpts, userGasLimit uint64, gasMultiplier float64, dryRun bool,
	simulateFn func(*bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction, error) {
	// Reset gas limit to user-specified value (0 = auto-estimate) before each transaction
	opts.GasLimit = userGasLimit

	// Simulate when dry-run is requested or when we need to apply a gas multiplier
	if dryRun || (gasMultiplier > 1.0 && userGasLimit == 0) {
		// Create a copy for simulation
		simulateOpts := *opts
		simulateOpts.NoSend = true

		simulatedTx, err := simulateFn(&simulateOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to simulate transaction: %w", err)
		}

		if gasMultiplier > 1.0 && userGasLimit == 0 {
			adjustedGas := uint64(float64(simulatedTx.Gas()) * gasMultiplier)
			opts.GasLimit = adjustedGas
			log.Info("Adjusted gas estimate", "original", simulatedTx.Gas(), "multiplier", gasMultiplier, "adjusted", adjustedGas)
		}

		return simulatedTx, nil
	}

	return nil, nil
}

func printDryRun(action string, tx *types.Transaction, from common.Address, gasOverride uint64) {
	gas := tx.Gas()
	if gasOverride > 0 {
		gas = gasOverride
	}

	logFields := []interface{}{
		"action", action,
		"from", from.Hex(),
	}
	if tx.To() != nil {
		logFields = append(logFields, "to", tx.To().Hex())
	}
	logFields = append(logFields, "value", tx.Value().String(), "estimatedGas", gas)

	if tx.Type() == types.DynamicFeeTxType {
		maxCost := new(big.Int).Mul(tx.GasFeeCap(), new(big.Int).SetUint64(gas))
		maxCostEth := new(big.Float).Quo(new(big.Float).SetInt(maxCost), new(big.Float).SetFloat64(1e18))
		logFields = append(logFields,
			"maxFee", tx.GasFeeCap().String(),
			"maxPriority", tx.GasTipCap().String(),
			"maxCostETH", maxCostEth.Text('f', 8),
		)
	} else {
		gasPrice := tx.GasPrice()
		cost := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gas))
		costEth := new(big.Float).Quo(new(big.Float).SetInt(cost), new(big.Float).SetFloat64(1e18))
		logFields = append(logFields,
			"gasPrice", gasPrice.String(),
			"estimatedCostETH", costEth.Text('f', 8),
		)
	}

	data := hex.EncodeToString(tx.Data())
	if len(data) > 128 {
		data = data[:128] + "..."
	}
	logFields = append(logFields, "txData", "0x"+data)

	log.Info("DRY RUN", logFields...)
}
