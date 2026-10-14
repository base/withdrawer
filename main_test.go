package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/base/withdrawer/signer"
	"github.com/base/withdrawer/withdraw"
	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

const testPrivateKey = "0000000000000000000000000000000000000000000000000000000000000001"

func TestCreateWithdrawHelperInitialRPCDeadline(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)
	s, err := signer.CreateSigner(testPrivateKey, "", "")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := CreateWithdrawHelper(ctx, server.URL, common.Hash{}, network{}, s, GasConfig{}, false)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("initial RPC query did not start")
	}
	<-ctx.Done()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("initial RPC query returned %v, want context deadline exceeded", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("initial RPC query did not exit after operation deadline")
	}
}

type rpcRequest struct {
	ID     json.RawMessage   `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

func localRPC(t *testing.T, respond func(rpcRequest) (any, bool)) (string, <-chan struct{}) {
	t.Helper()
	stalled := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		result, stall := respond(req)
		if stall {
			stalled <- struct{}{}
			select {
			case <-r.Context().Done():
			case <-release:
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(func() {
		close(release)
		server.Close()
	})
	return server.URL, stalled
}

func withdrawalReceipt(t *testing.T) *types.Receipt {
	t.Helper()
	contractABI, err := bindings.L2ToL1MessagePasserMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	event := contractABI.Events["MessagePassed"]
	data, err := event.Inputs.NonIndexed().Pack(big.NewInt(0), big.NewInt(21000), []byte{}, [32]byte{})
	if err != nil {
		t.Fatal(err)
	}
	return &types.Receipt{
		Status: types.ReceiptStatusSuccessful, BlockNumber: big.NewInt(1),
		Logs: []*types.Log{{Topics: []common.Hash{event.ID, {}, {}, {}}, Data: data}},
	}
}

func rpcResult(t *testing.T, req rpcRequest, receipt *types.Receipt) any {
	t.Helper()
	switch req.Method {
	case "eth_chainId":
		return "0x1"
	case "eth_getTransactionCount":
		return "0x0"
	case "eth_getTransactionReceipt":
		return receipt
	case "eth_getBlockByNumber":
		return &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(0), Extra: []byte{}}
	case "eth_getCode":
		return "0x01"
	}
	t.Errorf("unexpected RPC method %s", req.Method)
	return nil
}

func newTestWithdrawer(t *testing.T, ctx context.Context, l1, l2 string, faultProofs bool, gasLimit uint64, dryRun bool) withdraw.WithdrawHelper {
	t.Helper()
	s, err := signer.CreateSigner(testPrivateKey, "", "")
	if err != nil {
		t.Fatal(err)
	}
	w, err := CreateWithdrawHelper(ctx, l1, common.Hash{}, network{l2RPC: l2, faultProofs: faultProofs}, s, GasConfig{GasPrice: big.NewInt(1), GasLimit: gasLimit}, dryRun)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		switch w := w.(type) {
		case *withdraw.Withdrawer:
			w.L1Client.Close()
			w.L2Client.Close()
		case *withdraw.FPWithdrawer:
			w.L1Client.Close()
			w.L2Client.Close()
		}
	})
	return w
}

func assertRPCCanceled(t *testing.T, cancel context.CancelFunc, stalled <-chan struct{}, operation func() error) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- operation() }()
	select {
	case <-stalled:
	case err := <-done:
		t.Fatalf("operation returned before reaching stalled RPC: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("operation did not reach stalled RPC")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("operation returned %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("RPC did not exit after operation cancellation")
	}
}

func TestWithdrawHelpersRPCCancellation(t *testing.T) {
	for _, faultProofs := range []bool{false, true} {
		for _, operation := range []string{"receipt", "finalization-status", "proof-time", "provability", "proving", "upstream-code"} {
			t.Run(fmt.Sprintf("fault-proofs=%t/%s", faultProofs, operation), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				receipt := withdrawalReceipt(t)
				l2, l2Stalled := localRPC(t, func(req rpcRequest) (any, bool) {
					return rpcResult(t, req, receipt), operation == "receipt"
				})
				l1, stalled := localRPC(t, func(req rpcRequest) (any, bool) {
					if req.Method == "eth_call" {
						if operation == "upstream-code" {
							return "0x", false
						}
						if operation == "proving" && !faultProofs {
							var call struct{ Input hexutil.Bytes }
							if err := json.Unmarshal(req.Params[0], &call); err != nil {
								t.Error(err)
							}
							oracleABI, _ := bindings.L2OutputOracleMetaData.GetAbi()
							if len(call.Input) >= 4 && string(call.Input[:4]) == string(oracleABI.Methods["latestBlockNumber"].ID) {
								return hexutil.Encode(common.LeftPadBytes([]byte{1}, 32)), false
							}
						}
						return nil, true
					}
					if req.Method == "eth_getCode" && operation == "upstream-code" {
						return nil, true
					}
					return rpcResult(t, req, receipt), false
				})
				w := newTestWithdrawer(t, ctx, l1, l2, faultProofs, 0, false)
				if operation == "receipt" {
					stalled = l2Stalled
				}
				assertRPCCanceled(t, cancel, stalled, func() error {
					switch operation {
					case "receipt", "finalization-status":
						_, err := w.IsProofFinalized()
						return err
					case "proof-time":
						_, err := w.GetProvenWithdrawalTime()
						return err
					case "provability", "upstream-code":
						return w.CheckIfProvable()
					default:
						return w.ProveWithdrawal()
					}
				})
			})
		}
	}
}

func TestFinalizeWithdrawalRPCCancellation(t *testing.T) {
	for _, stage := range []string{"code", "estimate", "dry-run", "submit", "confirmation-rpc", "confirmation-poll"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			receipt := withdrawalReceipt(t)
			polled := make(chan struct{}, 1)
			l2, _ := localRPC(t, func(req rpcRequest) (any, bool) { return rpcResult(t, req, receipt), false })
			l1, stalled := localRPC(t, func(req rpcRequest) (any, bool) {
				switch req.Method {
				case "eth_call":
					return "0x", false
				case "eth_getCode":
					return "0x01", stage == "code"
				case "eth_estimateGas":
					return nil, true
				case "eth_sendRawTransaction":
					return common.Hash{}, stage == "submit"
				case "eth_getTransactionReceipt":
					if stage == "confirmation-poll" {
						polled <- struct{}{}
						return nil, false
					}
					return nil, true
				}
				return rpcResult(t, req, receipt), false
			})
			gasLimit := uint64(21000)
			if stage == "estimate" || stage == "dry-run" {
				gasLimit = 0
			}
			w := newTestWithdrawer(t, ctx, l1, l2, true, gasLimit, stage == "dry-run")
			if stage == "confirmation-poll" {
				stalled = polled
			}
			assertRPCCanceled(t, cancel, stalled, w.FinalizeWithdrawal)
		})
	}
}

func TestCreateWithdrawHelperDialDeadline(t *testing.T) {
	for _, layer := range []string{"L1", "L2"} {
		t.Run(layer, func(t *testing.T) {
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				select {
				case <-release:
				case <-r.Context().Done():
				}
			}))
			defer server.Close()
			defer close(release)
			endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
			l1 := endpoint
			if layer == "L2" {
				l1, _ = localRPC(t, func(req rpcRequest) (any, bool) { return rpcResult(t, req, nil), false })
			}
			s, err := signer.CreateSigner(testPrivateKey, "", "")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := CreateWithdrawHelper(ctx, l1, common.Hash{}, network{l2RPC: endpoint}, s, GasConfig{GasPrice: big.NewInt(1)}, false)
				done <- err
			}()
			select {
			case err := <-done:
				var timeoutErr net.Error
				if !errors.Is(err, context.DeadlineExceeded) && !(errors.As(err, &timeoutErr) && timeoutErr.Timeout()) {
					t.Fatalf("%s dial returned %v, want deadline or network timeout", layer, err)
				}
			case <-time.After(time.Second):
				t.Fatalf("%s dial did not exit on deadline", layer)
			}
		})
	}
}

type pendingNonceRPC struct {
	nonce atomic.Uint64
}

func (*pendingNonceRPC) ChainId() hexutil.Uint64 { return 1337 }

func (r *pendingNonceRPC) GetTransactionCount(_ common.Address, block rpc.BlockNumber) (hexutil.Uint64, error) {
	if block != rpc.PendingBlockNumber {
		return 0, fmt.Errorf("expected pending block, got %s", block)
	}
	return hexutil.Uint64(r.nonce.Load()), nil
}

func TestCreateWithdrawHelperUsesCurrentPendingNonce(t *testing.T) {
	s, err := signer.CreateSigner("0000000000000000000000000000000000000000000000000000000000000001", "", "")
	if err != nil {
		t.Fatal(err)
	}
	gas := GasConfig{GasLimit: 100000, GasPrice: big.NewInt(2_000_000_000), GasMultiplier: 1}
	for _, tc := range []struct {
		name        string
		faultProofs bool
	}{
		{name: "legacy"},
		{name: "fault-proofs", faultProofs: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := new(pendingNonceRPC)
			api.nonce.Store(7)
			rpcServer := rpc.NewServer()
			t.Cleanup(rpcServer.Stop)
			if err := rpcServer.RegisterName("eth", api); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(rpcServer)
			t.Cleanup(server.Close)
			helper, err := CreateWithdrawHelper(context.Background(), server.URL, common.Hash{}, network{
				l2RPC:         server.URL,
				portalAddress: "0x1234",
				faultProofs:   tc.faultProofs,
			}, s, gas, true)
			if err != nil {
				t.Fatal(err)
			}

			var opts *bind.TransactOpts
			var prove, finalize func(*bind.TransactOpts) (*types.Transaction, error)
			switch w := helper.(type) {
			case *withdraw.Withdrawer:
				t.Cleanup(w.L1Client.Close)
				t.Cleanup(w.L2Client.Close)
				opts = w.Opts
				withdrawal := bindings.TypesWithdrawalTransaction{Nonce: big.NewInt(0), Value: big.NewInt(0), GasLimit: big.NewInt(0)}
				prove = func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return w.Portal.ProveWithdrawalTransaction(opts, withdrawal, big.NewInt(0), bindings.TypesOutputRootProof{}, nil)
				}
				finalize = func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return w.Portal.FinalizeWithdrawalTransaction(opts, withdrawal)
				}
			case *withdraw.FPWithdrawer:
				t.Cleanup(w.L1Client.Close)
				t.Cleanup(w.L2Client.Close)
				opts = w.Opts
				withdrawal := bindingspreview.TypesWithdrawalTransaction{Nonce: big.NewInt(0), Value: big.NewInt(0), GasLimit: big.NewInt(0)}
				prove = func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return w.Portal.ProveWithdrawalTransaction(opts, withdrawal, big.NewInt(0), bindingspreview.TypesOutputRootProof{}, nil)
				}
				finalize = func(opts *bind.TransactOpts) (*types.Transaction, error) {
					return w.Portal.FinalizeWithdrawalTransaction(opts, withdrawal)
				}
			default:
				t.Fatalf("unexpected helper type %T", helper)
			}
			opts.NoSend = true
			for i, action := range []struct {
				name  string
				build func(*bind.TransactOpts) (*types.Transaction, error)
			}{
				{name: "prove", build: prove},
				{name: "finalize", build: finalize},
			} {
				pending := uint64(8 + i)
				api.nonce.Store(pending)
				tx, err := action.build(opts)
				if err != nil {
					t.Fatal(err)
				}
				if tx.Nonce() != pending {
					t.Fatalf("%s transaction nonce = %d, want current pending nonce %d (helper constructed at nonce 7)", action.name, tx.Nonce(), pending)
				}
				from, err := types.Sender(types.LatestSignerForChainID(big.NewInt(1337)), tx)
				if err != nil || from != s.Address() {
					t.Fatalf("%s transaction signer = %s, err = %v", action.name, from, err)
				}
				if tx.Gas() != gas.GasLimit || tx.GasPrice().Cmp(gas.GasPrice) != 0 {
					t.Fatalf("%s transaction gas = %d, gas price = %s", action.name, tx.Gas(), tx.GasPrice())
				}
			}
		})
	}
}

func TestCLIProcess(t *testing.T) {
	if os.Getenv("WITHDRAWER_TEST_MAIN") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"withdrawer"}, os.Args[i+1:]...)
			break
		}
	}
	main()
	os.Exit(0)
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestCLIProcess$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "WITHDRAWER_TEST_MAIN=1")
	output, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("CLI did not exit: %s", output)
	}
	return string(output), err
}

func TestTimeoutFlag(t *testing.T) {
	output, err := runCLI(t, "--help")
	if err != nil || !strings.Contains(output, "-timeout duration") || !strings.Contains(output, "default 10m0s") {
		t.Fatalf("timeout help missing or incorrect: %v\n%s", err, output)
	}
	output, err = runCLI(t, "--timeout=invalid")
	if err == nil || !strings.Contains(output, `invalid value "invalid" for flag -timeout`) {
		t.Fatalf("invalid duration was not rejected: %v\n%s", err, output)
	}
}

func TestCLIInitialRPCDeadline(t *testing.T) {
	l1, _ := localRPC(t, func(rpcRequest) (any, bool) { return nil, true })
	output, err := runCLI(t, "--timeout=100ms", "--rpc="+l1, "--withdrawal=0x1", "--private-key="+testPrivateKey, "--fault-proofs")
	if err == nil || !strings.Contains(output, "context deadline exceeded") {
		t.Fatalf("stalled initial RPC did not exit on deadline: %v\n%s", err, output)
	}
}

func TestCLIInitialRPCBackoffDeadline(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	output, err := runCLI(t, "--timeout=100ms", "--rpc="+server.URL, "--withdrawal=0x1", "--private-key="+testPrivateKey, "--fault-proofs")
	if err == nil || !strings.Contains(output, "context deadline exceeded") {
		t.Fatalf("initial RPC backoff did not exit on deadline: %v\n%s", err, output)
	}
	if requests.Load() != 1 {
		t.Fatalf("RPC attempts = %d, want only the initial attempt", requests.Load())
	}
}
