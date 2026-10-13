package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestRetryConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name      string
		config    RetryConfig
		wantCalls int32
		recover   bool
	}{
		{"disabled", RetryConfig{MaxRetries: 0, Delay: time.Millisecond}, 1, false},
		{"custom budget", RetryConfig{MaxRetries: 2, Delay: time.Millisecond}, 3, true},
		{"negative retries", RetryConfig{MaxRetries: -1, Delay: time.Millisecond}, 0, false},
		{"negative delay", RetryConfig{MaxRetries: 3, Delay: -1}, 0, false},
		{"zero delay", RetryConfig{MaxRetries: 0, Delay: 0}, 0, false},
		{"retry counter overflow", RetryConfig{MaxRetries: math.MaxInt, Delay: 1}, 0, false},
		{"duration overflow", RetryConfig{MaxRetries: 63, Delay: 2}, 0, false},
		{"largest representable delay", RetryConfig{MaxRetries: 1, Delay: math.MaxInt64}, 1, true},
		{"largest representable budget", RetryConfig{MaxRetries: 63, Delay: 1}, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
				attempt := calls.Add(1)
				if tc.name == "custom budget" && attempt == 3 || tc.wantCalls == 1 && tc.recover {
					return successfulReceipt(), nil, 0
				}
				return nil, nil, http.StatusTooManyRequests
			})
			err := waitForConfirmation(context.Background(), ethclient.NewClient(client), common.Hash{}, &tc.config)
			if (err == nil) != tc.recover || calls.Load() != tc.wantCalls {
				t.Fatalf("error %v after %d calls, want success %t after %d", err, calls.Load(), tc.recover, tc.wantCalls)
			}
		})
	}
}

func TestTypedRPCTransientErrors(t *testing.T) {
	for _, failure := range []rpcFailure{{-32002, "request timed out"}, {-32005, "rate limit exceeded"}} {
		t.Run(failure.Message, func(t *testing.T) {
			var calls atomic.Int32
			client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
				if calls.Add(1) == 1 {
					return nil, &failure, 0
				}
				return successfulReceipt(), nil, 0
			})
			config := &RetryConfig{MaxRetries: 1, Delay: time.Millisecond}
			if err := waitForConfirmation(context.Background(), ethclient.NewClient(client), common.Hash{}, config); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMissingOrInvalidWithdrawalDoesNotRetry(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, receipt := range []any{nil, successfulReceipt()} {
			var calls atomic.Int32
			client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) { calls.Add(1); return receipt, nil, 0 })
			var w WithdrawHelper = &Withdrawer{Ctx: context.Background(), L2Client: client}
			if fp {
				w = &FPWithdrawer{Ctx: context.Background(), L2Client: client}
			}
			_, err := w.GetProvenWithdrawalTime()
			if err == nil || calls.Load() != 1 {
				t.Fatalf("fp=%t invalid receipt: error %v after %d attempts", fp, err, calls.Load())
			}
		}
	}
}

func TestCanceledRPCDoesNotRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls atomic.Int32
	client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) { calls.Add(1); return nil, nil, http.StatusServiceUnavailable })
	err := waitForConfirmation(ctx, ethclient.NewClient(client), common.Hash{})
	if !errors.Is(err, context.Canceled) || calls.Load() != 0 {
		t.Fatalf("canceled operation: %v after %d calls", err, calls.Load())
	}
}

func TestPendingContractReadsRecover(t *testing.T) {
	var calls atomic.Int32
	client := localRPC(t, func(req rpcRequest) (any, *rpcFailure, int) {
		if calls.Add(1) == 1 {
			return nil, nil, http.StatusServiceUnavailable
		}
		var block string
		if err := json.Unmarshal(req.Params[1], &block); err != nil {
			t.Error(err)
		}
		var data [32]byte
		if block == "pending" {
			data[31] = 1
		}
		return hexutil.Encode(data[:]), nil, 0
	})
	config := &RetryConfig{MaxRetries: 1, Delay: time.Millisecond}
	backend := NewRetryingBackend(context.Background(), ethclient.NewClient(client), config)
	portal, err := bindings.NewOptimismPortal(common.Address{}, backend)
	if err != nil {
		t.Fatal(err)
	}
	finalized, err := portal.FinalizedWithdrawals(&bind.CallOpts{Pending: true}, common.Hash{})
	if err != nil || !finalized {
		t.Fatalf("pending read returned %t: %v", finalized, err)
	}
}

type canceledReadRPC struct{}

func (c canceledReadRPC) GetTransactionReceipt(context.Context, common.Hash) (*types.Receipt, error) {
	return successfulReceipt(), nil
}

func (c canceledReadRPC) Call(ctx context.Context, _ map[string]any, _ string) (hexutil.Bytes, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestTransitiveReadUsesWithdrawerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := rpc.NewServer()
	if err := server.RegisterName("eth", canceledReadRPC{}); err != nil {
		t.Fatal(err)
	}
	defer server.Stop()
	client := rpc.DialInProc(server)
	defer client.Close()
	backend := NewRetryingBackend(ctx, ethclient.NewClient(client), nil)
	portal, err := bindingspreview.NewOptimismPortal2(common.Address{}, backend)
	if err != nil {
		t.Fatal(err)
	}
	factory, err := bindings.NewDisputeGameFactory(common.Address{}, backend)
	if err != nil {
		t.Fatal(err)
	}
	w := FPWithdrawer{Ctx: ctx, L2Client: client, Portal: portal, Factory: factory}
	timer := time.AfterFunc(50*time.Millisecond, cancel)
	defer timer.Stop()
	started := time.Now()
	err = w.CheckIfProvable()
	if !errors.Is(err, context.Canceled) || time.Since(started) > time.Second {
		t.Fatalf("transitive read cancellation: %v in %v", err, time.Since(started))
	}
}
