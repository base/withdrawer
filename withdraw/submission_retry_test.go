package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestFinalizeReconcilesExactSubmission(t *testing.T) {
	for _, tc := range []struct {
		name            string
		receipt         string
		knownTx         string
		retries         int
		ambiguous       bool
		canceled        bool
		receiptFailures int
		lookupFailures  int
		cancelQuery     string
		wantReceipts    int
		wantLookups     int
		wantError       string
	}{
		{name: "mined before resend", receipt: "exact", retries: 1, ambiguous: true, wantReceipts: 2},
		{name: "mined but reverted", receipt: "failed", retries: 1, ambiguous: true, wantReceipts: 2, wantError: "unsuccessful withdrawal receipt status"},
		{name: "exact transaction found", receipt: "exact", knownTx: "exact", retries: 1, ambiguous: true, wantReceipts: 2, wantLookups: 1},
		{name: "absent transaction", retries: 1, ambiguous: true, wantReceipts: 1, wantLookups: 1, wantError: "nonce too low"},
		{name: "different receipt and transaction", receipt: "different", knownTx: "different", retries: 1, ambiguous: true, wantReceipts: 1, wantLookups: 1, wantError: "nonce too low"},
		{name: "nonce too low without ambiguity", receipt: "exact", retries: 1, wantError: "nonce too low"},
		{name: "exhausted but mined", receipt: "exact", ambiguous: true, wantReceipts: 2},
		{name: "exhausted and absent", ambiguous: true, wantReceipts: 1, wantLookups: 1, wantError: "503"},
		{name: "canceled ambiguity", receipt: "exact", retries: 1, ambiguous: true, canceled: true, wantError: "context canceled"},
		{name: "transient receipt query", receipt: "exact", retries: 1, ambiguous: true, receiptFailures: 1, wantReceipts: 3},
		{name: "transient transaction lookup", receipt: "exact", knownTx: "exact", retries: 1, ambiguous: true, lookupFailures: 1, wantReceipts: 2, wantLookups: 2},
		{name: "both queries transient", receipt: "exact", knownTx: "exact", retries: 1, ambiguous: true, receiptFailures: 1, lookupFailures: 1, wantReceipts: 3, wantLookups: 2},
		{name: "different hashes after transient queries", receipt: "different", knownTx: "different", retries: 1, ambiguous: true, receiptFailures: 1, lookupFailures: 1, wantReceipts: 2, wantLookups: 2, wantError: "nonce too low"},
		{name: "queries exhausted", retries: 1, ambiguous: true, receiptFailures: 2, lookupFailures: 2, wantReceipts: 2, wantLookups: 2, wantError: "nonce too low"},
		{name: "receipt exhausted but transaction found", receipt: "exact", knownTx: "exact", retries: 1, ambiguous: true, receiptFailures: 2, wantReceipts: 3, wantLookups: 1},
		{name: "cancel receipt backoff", retries: 1, ambiguous: true, receiptFailures: 1, cancelQuery: "receipt", wantReceipts: 1, wantError: "context canceled"},
		{name: "cancel transaction lookup backoff", retries: 1, ambiguous: true, lookupFailures: 1, cancelQuery: "lookup", wantReceipts: 1, wantLookups: 1, wantError: "context canceled"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			l2Receipt := messageReceipt(t)
			key, err := crypto.GenerateKey()
			if err != nil {
				t.Fatal(err)
			}
			opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1))
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx = rpc.NewContextWithHeaders(ctx, http.Header{"Authorization": {"Bearer withdrawer"}, "X-Rpc-Route": {"withdrawer"}})
			optsCtx, cancelOpts := context.WithCancel(context.Background())
			defer cancelOpts()
			optsCtx = rpc.NewContextWithHeaders(optsCtx, http.Header{"Authorization": {"Bearer transaction"}, "X-Rpc-Route": {"transaction"}})
			opts.Context, opts.Nonce, opts.GasPrice = optsCtx, big.NewInt(7), big.NewInt(1)
			signer := opts.Signer
			signed := 0
			opts.Signer = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
				signed++
				return signer(address, tx)
			}
			var mu sync.Mutex
			var payloads []string
			var submitted types.Transaction
			receiptQueries := 0
			lookupQueries := 0
			client := localRPC(t, func(req rpcRequest) (any, *rpcFailure, int) {
				mu.Lock()
				defer mu.Unlock()
				if req.Method == "eth_sendRawTransaction" || req.Method == "eth_getTransactionByHash" || req.Method == "eth_getTransactionReceipt" && string(req.Params[0]) != `"`+(common.Hash{}).Hex()+`"` {
					wantRoute := "transaction"
					if req.Method == "eth_getTransactionReceipt" && receiptQueries >= min(tc.receiptFailures+1, tc.retries+1) {
						wantRoute = "withdrawer"
					}
					if req.Header.Get("Authorization") != "Bearer "+wantRoute || req.Header.Get("X-Rpc-Route") != wantRoute {
						t.Errorf("%s headers %v, want %s context headers", req.Method, req.Header, wantRoute)
					}
				}
				switch req.Method {
				case "eth_getCode":
					return "0x01", nil, 0
				case "eth_call":
					return "0x", nil, 0
				case "eth_getTransactionReceipt", "eth_getTransactionByHash":
					var hash common.Hash
					if err := json.Unmarshal(req.Params[0], &hash); err != nil {
						t.Error(err)
					}
					if hash == (common.Hash{}) && req.Method == "eth_getTransactionReceipt" {
						return l2Receipt, nil, 0
					}
					if len(payloads) == 0 || hash != submitted.Hash() {
						t.Errorf("queried different submission hash %s", hash)
					}
					if req.Method == "eth_getTransactionByHash" {
						lookupQueries++
						if lookupQueries <= tc.lookupFailures {
							if tc.cancelQuery == "lookup" {
								time.AfterFunc(10*time.Millisecond, cancelOpts)
							}
							return nil, nil, http.StatusServiceUnavailable
						}
						if tc.knownTx == "exact" {
							return &submitted, nil, 0
						}
						if tc.knownTx == "different" {
							return types.NewTx(&types.LegacyTx{Nonce: 8, GasPrice: big.NewInt(1), V: big.NewInt(27), R: big.NewInt(1), S: big.NewInt(1)}), nil, 0
						}
						return nil, nil, 0
					}
					receiptQueries++
					if receiptQueries <= tc.receiptFailures {
						if tc.cancelQuery == "receipt" {
							time.AfterFunc(10*time.Millisecond, cancelOpts)
						}
						return nil, nil, http.StatusServiceUnavailable
					}
					if tc.receipt == "" || tc.knownTx == "exact" && tc.receiptFailures <= tc.retries && receiptQueries == tc.receiptFailures+1 {
						return nil, nil, 0
					}
					receipt := successfulReceipt()
					receipt.TxHash = submitted.Hash()
					if tc.receipt == "different" {
						receipt.TxHash = common.HexToHash("0x99")
					}
					if tc.receipt == "failed" {
						receipt.Status = types.ReceiptStatusFailed
					}
					return receipt, nil, 0
				case "eth_sendRawTransaction":
					var raw string
					if err := json.Unmarshal(req.Params[0], &raw); err != nil {
						t.Error(err)
					}
					payloads = append(payloads, raw)
					if len(payloads) == 1 {
						data, err := hexutil.Decode(raw)
						if err != nil {
							t.Error(err)
						}
						if err := submitted.UnmarshalBinary(data); err != nil {
							t.Error(err)
						}
						if tc.ambiguous {
							if tc.canceled {
								cancel()
							}
							return nil, nil, http.StatusServiceUnavailable
						}
					}
					if raw != payloads[0] {
						t.Error("resend changed signed transaction")
					}
					return nil, &rpcFailure{-32000, "nonce too low"}, 0
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
			delay := time.Millisecond
			if tc.cancelQuery != "" {
				delay = 100 * time.Millisecond
			}
			w := FPWithdrawer{Ctx: ctx, L1Client: l1, L2Client: client, Portal: portal, Opts: opts, UserGasLimit: 65536, GasMultiplier: 1, Retry: &RetryConfig{MaxRetries: tc.retries, Delay: delay}}
			err = w.FinalizeWithdrawal()
			if tc.wantError == "" && err != nil || tc.wantError != "" && (err == nil || !strings.Contains(err.Error(), tc.wantError)) {
				t.Fatalf("finalize error %v, want %q", err, tc.wantError)
			}
			if (tc.canceled || tc.cancelQuery != "") && !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
			mu.Lock()
			defer mu.Unlock()
			wantSends := 1
			if tc.ambiguous && tc.retries > 0 && !tc.canceled {
				wantSends = 2
			}
			if signed != 1 || len(payloads) != wantSends || receiptQueries != tc.wantReceipts || lookupQueries != tc.wantLookups {
				t.Fatalf("signed %d, submissions %d, receipt queries %d, transaction lookups %d", signed, len(payloads), receiptQueries, lookupQueries)
			}
		})
	}
}
