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
			fmt.Printf("waiting for tx confirmation\n")
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
	fmt.Printf("%s confirmed\n", tx.String())
	return nil
}

// prepareGasOpts resets the gas limit and applies gas multiplier if needed.
// It returns a function that should be called to estimate gas and apply the multiplier.
// The estimateFn should perform a NoSend transaction and return the estimated gas.
func prepareGasOpts(opts *bind.TransactOpts, userGasLimit uint64, gasMultiplier float64, estimateFn func(*bind.TransactOpts) (uint64, error)) error {
	// Reset gas limit to user-specified value (0 = auto-estimate) before each transaction
	opts.GasLimit = userGasLimit

	// Apply gas multiplier if set and no explicit gas limit
	if gasMultiplier > 1.0 && userGasLimit == 0 {
		// Create a copy for estimation
		estimateOpts := *opts
		estimateOpts.NoSend = true

		estimatedGas, err := estimateFn(&estimateOpts)
		if err != nil {
			return fmt.Errorf("failed to estimate gas: %w", err)
		}

		adjustedGas := uint64(float64(estimatedGas) * gasMultiplier)
		opts.GasLimit = adjustedGas
		log.Info("Adjusted gas estimate", "original", estimatedGas, "multiplier", gasMultiplier, "adjusted", adjustedGas)
	}

	return nil
}

func printDryRun(action string, tx *types.Transaction, from common.Address) {
	fmt.Println("=== DRY RUN ===")
	fmt.Printf("Action:         %s\n", action)
	fmt.Printf("From:           %s\n", from.Hex())
	if tx.To() != nil {
		fmt.Printf("To:             %s\n", tx.To().Hex())
	}
	fmt.Printf("Estimated Gas:  %d\n", tx.Gas())

	if tx.Type() == types.DynamicFeeTxType {
		fmt.Printf("Max Fee:        %s wei\n", tx.GasFeeCap().String())
		fmt.Printf("Max Priority:   %s wei\n", tx.GasTipCap().String())
		maxCost := new(big.Int).Mul(tx.GasFeeCap(), new(big.Int).SetUint64(tx.Gas()))
		maxCostEth := new(big.Float).Quo(new(big.Float).SetInt(maxCost), new(big.Float).SetFloat64(1e18))
		fmt.Printf("Max Cost:       %s ETH\n", maxCostEth.Text('f', 8))
	} else {
		gasPrice := tx.GasPrice()
		fmt.Printf("Gas Price:      %s wei\n", gasPrice.String())
		cost := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(tx.Gas()))
		costEth := new(big.Float).Quo(new(big.Float).SetInt(cost), new(big.Float).SetFloat64(1e18))
		fmt.Printf("Estimated Cost: %s ETH\n", costEth.Text('f', 8))
	}

	data := hex.EncodeToString(tx.Data())
	if len(data) > 128 {
		data = data[:128] + "..."
	}
	fmt.Printf("Tx Data:        0x%s\n", data)
	fmt.Println("===============")
}
