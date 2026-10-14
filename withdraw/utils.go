package withdraw

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 2 * time.Second
)

// RetryConfig sets the retry budget after the initial RPC call.
type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
}

func (r RetryConfig) Validate() error {
	if r.MaxRetries < 0 || r.MaxRetries > 63 {
		return errors.New("max-retries must be between 0 and 63")
	}
	if r.Delay <= 0 {
		return errors.New("retry-delay must be positive")
	}
	if r.MaxRetries > 0 && r.Delay > time.Duration(math.MaxInt64>>(r.MaxRetries-1)) {
		return errors.New("retry-delay exponential backoff overflows time.Duration")
	}
	return nil
}

// Do retries only transient RPC failures; nil uses the CLI defaults.
func (r *RetryConfig) Do(ctx context.Context, operation string, call func() error) error {
	config := RetryConfig{MaxRetries: DefaultMaxRetries, Delay: DefaultRetryDelay}
	if r != nil {
		config = *r
	}
	if err := config.Validate(); err != nil {
		return err
	}
	delay := config.Delay
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := call()
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt == config.MaxRetries || !isTransientRPCError(err) {
			return err
		}
		log.Warn("Retrying RPC operation", "operation", operation, "attempt", attempt+2, "budget", config.MaxRetries+1, "error", err, "delay", delay)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if attempt+1 < config.MaxRetries {
			delay *= 2
		}
	}
}

func retryRPC[T any](ctx context.Context, config *RetryConfig, operation string, call func() (T, error)) (T, error) {
	var result T
	err := config.Do(ctx, operation, func() error {
		var err error
		result, err = call()
		return err
	})
	return result, err
}

func isTransientRPCError(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	var rpcErr rpc.Error
	if errors.As(err, &rpcErr) {
		var dataErr rpc.DataError
		if errors.As(err, &dataErr) && dataErr.ErrorData() != nil {
			return false
		}
		message := strings.ToLower(strings.TrimSpace(rpcErr.Error()))
		switch rpcErr.ErrorCode() {
		case -32002:
			return message == "request timed out"
		case -32005:
			return message == "rate limit exceeded" || message == "request limit exceeded"
		}
		return false
	}
	var httpErr rpc.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case 408, 429, 502, 503, 504:
			return true
		}
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr != context.DeadlineExceeded && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseServiceRestart, websocket.CloseTryAgainLater:
			return true
		}
		return false
	}
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE)
}

func sendTransaction(ctx, optsCtx context.Context, client *ethclient.Client, config *RetryConfig, operation string, tx *types.Transaction) (err error) {
	log.Info("Prepared signed withdrawal transaction for submission", "operation", operation, "l1TxHash", tx.Hash())
	if optsCtx != nil {
		if err := optsCtx.Err(); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		withdrawerCtx := ctx
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(optsCtx)
		stop := context.AfterFunc(withdrawerCtx, cancel)
		defer cancel()
		defer stop()
		// WithCancel cannot report a deadline from the other context.
		defer func() {
			if optsCtx.Err() != nil {
				err = optsCtx.Err()
			} else if withdrawerCtx.Err() != nil {
				err = withdrawerCtx.Err()
			}
		}()
	}
	ambiguous := false
	err = config.Do(ctx, operation, func() error {
		err := client.SendTransaction(ctx, tx)
		if isTransientRPCError(err) {
			ambiguous = true
		}
		var rpcErr rpc.Error
		if errors.As(err, &rpcErr) && rpcErr.ErrorCode() == -32000 {
			var dataErr rpc.DataError
			if errors.As(err, &dataErr) && dataErr.ErrorData() != nil {
				return err
			}
			message := strings.ToLower(strings.TrimSpace(rpcErr.Error()))
			if message == "already known" || message == "known transaction" {
				return nil
			}
		}
		return err
	})
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err == nil || !ambiguous {
		return err
	}
	// A resend can report nonce too low after the exact transaction has mined.
	receipt, receiptErr := retryRPC(ctx, config, "eth_getTransactionReceipt.reconcile", func() (*types.Receipt, error) {
		return client.TransactionReceipt(ctx, tx.Hash())
	})
	if receiptErr == nil && receipt.TxHash == tx.Hash() && ctx.Err() == nil {
		return nil
	}
	knownTx, txErr := retryRPC(ctx, config, "eth_getTransactionByHash.reconcile", func() (*types.Transaction, error) {
		knownTx, _, err := client.TransactionByHash(ctx, tx.Hash())
		return knownTx, err
	})
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if txErr == nil && knownTx.Hash() == tx.Hash() {
		return nil
	}
	return err
}

// NewRetryingBackend retries binding reads and preparation, not submission.
func NewRetryingBackend(ctx context.Context, client *ethclient.Client, config *RetryConfig) *retryingBackend {
	return &retryingBackend{Client: client, ctx: ctx, config: config}
}

type retryingBackend struct {
	*ethclient.Client
	ctx    context.Context
	config *RetryConfig
}

func backendRPC[T any](b *retryingBackend, ctx context.Context, operation string, call func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(b.ctx, cancel)
	defer stop()
	defer cancel()
	if err := b.ctx.Err(); err != nil {
		cancel()
	}
	return retryRPC(ctx, b.config, operation, func() (T, error) { return call(ctx) })
}

func (b *retryingBackend) CallContract(ctx context.Context, msg ethereum.CallMsg, block *big.Int) ([]byte, error) {
	return backendRPC(b, ctx, "eth_call", func(ctx context.Context) ([]byte, error) { return b.Client.CallContract(ctx, msg, block) })
}

func (b *retryingBackend) PendingCallContract(ctx context.Context, msg ethereum.CallMsg) ([]byte, error) {
	return backendRPC(b, ctx, "eth_call.pending", func(ctx context.Context) ([]byte, error) { return b.Client.PendingCallContract(ctx, msg) })
}

func (b *retryingBackend) CallContractAtHash(ctx context.Context, msg ethereum.CallMsg, hash common.Hash) ([]byte, error) {
	return backendRPC(b, ctx, "eth_call.blockHash", func(ctx context.Context) ([]byte, error) { return b.Client.CallContractAtHash(ctx, msg, hash) })
}

func (b *retryingBackend) CodeAt(ctx context.Context, address common.Address, block *big.Int) ([]byte, error) {
	return backendRPC(b, ctx, "eth_getCode", func(ctx context.Context) ([]byte, error) { return b.Client.CodeAt(ctx, address, block) })
}

func (b *retryingBackend) CodeAtHash(ctx context.Context, address common.Address, hash common.Hash) ([]byte, error) {
	return backendRPC(b, ctx, "eth_getCode.blockHash", func(ctx context.Context) ([]byte, error) { return b.Client.CodeAtHash(ctx, address, hash) })
}

func (b *retryingBackend) PendingCodeAt(ctx context.Context, address common.Address) ([]byte, error) {
	return backendRPC(b, ctx, "eth_getCode.pending", func(ctx context.Context) ([]byte, error) { return b.Client.PendingCodeAt(ctx, address) })
}

func (b *retryingBackend) PendingNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	return backendRPC(b, ctx, "eth_getTransactionCount.pending", func(ctx context.Context) (uint64, error) { return b.Client.PendingNonceAt(ctx, address) })
}

func (b *retryingBackend) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return backendRPC(b, ctx, "eth_gasPrice", b.Client.SuggestGasPrice)
}

func (b *retryingBackend) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	return backendRPC(b, ctx, "eth_maxPriorityFeePerGas", b.Client.SuggestGasTipCap)
}

func (b *retryingBackend) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return backendRPC(b, ctx, "eth_estimateGas", func(ctx context.Context) (uint64, error) { return b.Client.EstimateGas(ctx, msg) })
}

func (b *retryingBackend) HeaderByNumber(ctx context.Context, block *big.Int) (*types.Header, error) {
	return backendRPC(b, ctx, "eth_getBlockByNumber", func(ctx context.Context) (*types.Header, error) { return b.Client.HeaderByNumber(ctx, block) })
}

type retryingL2 struct {
	*ethclient.Client
	proof  *gethclient.Client
	config *RetryConfig
}

func (c retryingL2) TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	return retryRPC(ctx, c.config, "eth_getTransactionReceipt.L2", func() (*types.Receipt, error) { return c.Client.TransactionReceipt(ctx, hash) })
}

func (c retryingL2) HeaderByNumber(ctx context.Context, block *big.Int) (*types.Header, error) {
	return retryRPC(ctx, c.config, "eth_getBlockByNumber.L2", func() (*types.Header, error) { return c.Client.HeaderByNumber(ctx, block) })
}

func (c retryingL2) GetProof(ctx context.Context, address common.Address, keys []string, block *big.Int) (*gethclient.AccountResult, error) {
	return retryRPC(ctx, c.config, "eth_getProof.L2", func() (*gethclient.AccountResult, error) { return c.proof.GetProof(ctx, address, keys, block) })
}

type WithdrawHelper interface {
	CheckIfProvable() error
	GetProvenWithdrawalTime() (uint64, error)
	ProveWithdrawal() error
	IsProofFinalized() (bool, error)
	FinalizeWithdrawal() error
}

func txBlock(ctx context.Context, l2c *rpc.Client, l2TxHash common.Hash, config *RetryConfig) (*big.Int, error) {
	l2 := retryingL2{Client: ethclient.NewClient(l2c), config: config}
	// Figure out when our withdrawal was included
	receipt, err := l2.TransactionReceipt(ctx, l2TxHash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch receipt for L2 withdrawal transaction %s: %w", l2TxHash, err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("L2 withdrawal transaction failed: tx %s, receipt status %d", l2TxHash, receipt.Status)
	}
	return receipt.BlockNumber, nil
}

func waitForConfirmation(ctx context.Context, client *ethclient.Client, tx common.Hash, operation string, configs ...*RetryConfig) error {
	var config *RetryConfig
	if len(configs) > 0 {
		config = configs[0]
	}
	for {
		receipt, err := retryRPC(ctx, config, "eth_getTransactionReceipt.L1", func() (*types.Receipt, error) { return client.TransactionReceipt(ctx, tx) })
		if err == ethereum.NotFound {
			log.Info("Waiting for tx confirmation", "txHash", tx.String())
			select {
			case <-ctx.Done():
				return fmt.Errorf("waiting for confirmation of L1 %s transaction %s: %w", operation, tx, ctx.Err())
			case <-time.After(5 * time.Second):
			}
		} else if err != nil {
			return fmt.Errorf("failed to fetch receipt for L1 %s transaction %s: %w", operation, tx, err)
		} else if receipt.Status != types.ReceiptStatusSuccessful {
			return fmt.Errorf("L1 %s transaction failed: tx %s, receipt status %d", operation, tx, receipt.Status)
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
