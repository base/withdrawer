package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type rpcRequest struct {
	ID     json.RawMessage   `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
	Header http.Header       `json:"-"`
}

type rpcFailure struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func localRPC(t *testing.T, handle func(rpcRequest) (any, *rpcFailure, int)) *rpc.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			return
		}
		req.Header = r.Header
		result, failure, status := handle(req)
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		response := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if failure != nil {
			response["error"] = failure
		} else {
			response["result"] = result
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := rpc.Dial(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	return client
}

func successfulReceipt() *types.Receipt {
	return &types.Receipt{Status: 1, BlockNumber: big.NewInt(1), Logs: []*types.Log{}, TxHash: common.HexToHash("0x01")}
}

func TestReceiptDefaultRetriesAndBackoff(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	var times []time.Time
	client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
		mu.Lock()
		defer mu.Unlock()
		times = append(times, time.Now())
		if len(times) < 4 {
			return nil, nil, http.StatusServiceUnavailable
		}
		return successfulReceipt(), nil, 0
	})
	err := waitForConfirmation(context.Background(), ethclient.NewClient(client), common.Hash{})
	if err != nil {
		t.Fatalf("transient receipt RPC must recover with default budget: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(times) != 4 {
		t.Fatalf("got %d attempts, want 4", len(times))
	}
	for i, delay := range []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second} {
		elapsed := times[i+1].Sub(times[i])
		if elapsed < delay || elapsed > delay+time.Second {
			t.Errorf("wait %d = %v, want %v", i, elapsed, delay)
		}
	}
}

func TestReceiptExhaustion(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	attempts := 0
	client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
		mu.Lock()
		defer mu.Unlock()
		attempts++
		return nil, nil, http.StatusServiceUnavailable
	})
	err := waitForConfirmation(context.Background(), ethclient.NewClient(client), common.Hash{})
	mu.Lock()
	defer mu.Unlock()
	var httpErr rpc.HTTPError
	if !errors.As(err, &httpErr) || attempts != 4 {
		t.Fatalf("exhaustion = %v after %d attempts, want original HTTP error after 4", err, attempts)
	}
}

func TestReceiptPermanentErrorStopsRetry(t *testing.T) {
	for _, failure := range []rpcFailure{
		{-32000, "insufficient funds for gas * price + value"}, {-32000, "execution reverted"},
		{-32602, "invalid params"}, {4001, "user rejected request"}, {-32000, "withdrawal not found"},
		{-32603, "internal error"}, {-32000, "nonce too low"}, {-32005, "execution reverted"},
	} {
		t.Run(failure.Message, func(t *testing.T) {
			t.Parallel()
			var mu sync.Mutex
			attempts := 0
			client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
				mu.Lock()
				defer mu.Unlock()
				attempts++
				if attempts == 1 {
					return nil, nil, http.StatusServiceUnavailable
				}
				return nil, &failure, 0
			})
			err := waitForConfirmation(context.Background(), ethclient.NewClient(client), common.Hash{})
			mu.Lock()
			defer mu.Unlock()
			var rpcErr rpc.Error
			if !errors.As(err, &rpcErr) || rpcErr.ErrorCode() != failure.Code || attempts != 2 {
				t.Fatalf("permanent failure = %v after %d attempts, want %q after 2", err, attempts, failure.Message)
			}
		})
	}
}

func TestCancellationDuringBackoff(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempts := 0
	client := localRPC(t, func(rpcRequest) (any, *rpcFailure, int) {
		mu.Lock()
		defer mu.Unlock()
		attempts++
		time.AfterFunc(50*time.Millisecond, cancel)
		return nil, nil, http.StatusServiceUnavailable
	})
	started := time.Now()
	err := waitForConfirmation(ctx, ethclient.NewClient(client), common.Hash{})
	mu.Lock()
	defer mu.Unlock()
	if !errors.Is(err, context.Canceled) || attempts != 1 || time.Since(started) > time.Second {
		t.Fatalf("backoff cancellation = %v after %d attempts in %v", err, attempts, time.Since(started))
	}
}

func messageReceipt(t *testing.T) *types.Receipt {
	t.Helper()
	abi, err := bindings.L2ToL1MessagePasserMetaData.GetAbi()
	if err != nil {
		t.Fatal(err)
	}
	event := abi.Events["MessagePassed"]
	data, err := event.Inputs.NonIndexed().Pack(big.NewInt(0), big.NewInt(21000), []byte{}, [32]byte{})
	if err != nil {
		t.Fatal(err)
	}
	receipt := successfulReceipt()
	receipt.Logs = []*types.Log{{Topics: []common.Hash{event.ID, {}, common.HexToHash("0x1234"), common.HexToHash("0x5678")}, Data: data}}
	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := withdrawals.WithdrawalHash(ev)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Logs[0].Data, err = event.Inputs.NonIndexed().Pack(ev.Value, ev.GasLimit, ev.Data, [32]byte(hash))
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func TestFinalizeRebroadcastsIdenticalSignedTransaction(t *testing.T) {
	for _, response := range []rpcFailure{
		{-32000, "already known"}, {-32000, "known transaction"},
		{-32000, "invalid transaction: already known"}, {-32000, "insufficient funds"},
		{-32602, "already known"},
	} {
		t.Run(response.Message, func(t *testing.T) {
			t.Parallel()
			testFinalizeRebroadcast(t, response)
		})
	}
}

func testFinalizeRebroadcast(t *testing.T, response rpcFailure) {
	t.Helper()
	receipt := messageReceipt(t)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	opts.Context = context.Background()
	var mu sync.Mutex
	var payloads []string
	client := localRPC(t, func(req rpcRequest) (any, *rpcFailure, int) {
		mu.Lock()
		defer mu.Unlock()
		switch req.Method {
		case "eth_getTransactionReceipt":
			return receipt, nil, 0
		case "eth_call":
			return "0x", nil, 0
		case "eth_getCode":
			return "0x01", nil, 0
		case "eth_getTransactionByHash":
			return nil, nil, 0
		case "eth_getBlockByNumber":
			return &types.Header{Number: big.NewInt(1), Difficulty: big.NewInt(0), BaseFee: big.NewInt(10)}, nil, 0
		case "eth_maxPriorityFeePerGas":
			return "0x1", nil, 0
		case "eth_estimateGas":
			return "0x10000", nil, 0
		case "eth_getTransactionCount":
			if len(payloads) > 0 {
				return nil, &rpcFailure{-32000, "nonce queried after ambiguous submission"}, 0
			}
			return "0x7", nil, 0
		case "eth_sendRawTransaction":
			var raw string
			if err := json.Unmarshal(req.Params[0], &raw); err != nil {
				t.Error(err)
			}
			payloads = append(payloads, raw)
			if len(payloads) == 1 {
				return nil, nil, http.StatusServiceUnavailable
			}
			if raw != payloads[0] {
				return nil, &rpcFailure{-32000, "different transaction after ambiguous submission"}, 0
			}
			var tx types.Transaction
			bytes, err := hexutil.Decode(raw)
			if err != nil {
				t.Error(err)
			}
			if err := tx.UnmarshalBinary(bytes); err != nil {
				t.Error(err)
			}
			sender, err := types.Sender(types.LatestSignerForChainID(big.NewInt(1)), &tx)
			if err != nil || sender != opts.From || tx.Nonce() != 7 {
				t.Errorf("invalid signed transaction: sender %s nonce %d error %v", sender, tx.Nonce(), err)
			}
			return nil, &response, 0
		default:
			t.Errorf("unexpected RPC %s", req.Method)
			return nil, &rpcFailure{-32601, "unsupported"}, 0
		}
	})
	l1 := ethclient.NewClient(client)
	portal, err := bindingspreview.NewOptimismPortal2(common.HexToAddress("0x01"), l1)
	if err != nil {
		t.Fatal(err)
	}
	w := FPWithdrawer{Ctx: context.Background(), L1Client: l1, L2Client: client, Portal: portal, Opts: opts, GasMultiplier: 1}
	err = w.FinalizeWithdrawal()
	known := response.Code == -32000 && (response.Message == "already known" || response.Message == "known transaction")
	if known && err != nil {
		t.Fatalf("ambiguous submission must recover without rebuilding: %v", err)
	}
	if !known {
		var rpcErr rpc.Error
		if !errors.As(err, &rpcErr) || rpcErr.ErrorCode() != response.Code || rpcErr.Error() != response.Message {
			t.Fatalf("invalid rebroadcast error was hidden: %v", err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(payloads) != 2 || payloads[0] != payloads[1] {
		t.Fatalf("expected identical rebroadcast, got %d payloads", len(payloads))
	}
}
