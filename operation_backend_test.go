package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestOperationBackendContexts(t *testing.T) {
	for _, method := range []string{"eth_call", "eth_getCode"} {
		for _, scenario := range []string{"caller-deadline", "caller-cancel", "caller-cancel-with-deadline-cause", "caller-headers", "operation-deadline", "operation-cancel", "operation-already-canceled", "operation-already-expired"} {
			t.Run(method+"/"+scenario, func(t *testing.T) {
				operation, cancelOperation := context.WithTimeout(context.Background(), 3*time.Second)
				t.Cleanup(cancelOperation)
				caller := context.Background()
				cancel := cancelOperation
				wantErr := error(context.Canceled)
				alreadyDone := false
				switch scenario {
				case "caller-deadline":
					var cancelCaller context.CancelFunc
					caller, cancelCaller = context.WithTimeout(caller, 150*time.Millisecond)
					t.Cleanup(cancelCaller)
					wantErr = context.DeadlineExceeded
				case "caller-cancel":
					caller, cancel = context.WithCancel(caller)
					t.Cleanup(cancel)
				case "caller-cancel-with-deadline-cause":
					var cancelCaller context.CancelCauseFunc
					caller, cancelCaller = context.WithCancelCause(caller)
					wantErr = context.DeadlineExceeded
					cancel = func() { cancelCaller(context.DeadlineExceeded) }
					t.Cleanup(cancel)
				case "caller-headers":
					caller = rpc.NewContextWithHeaders(caller, http.Header{"X-Caller": {"preserved"}})
					wantErr = nil
				case "operation-deadline", "operation-already-expired":
					timeout := 150 * time.Millisecond
					if scenario == "operation-already-expired" {
						timeout = -time.Second
						alreadyDone = true
					}
					operation, cancel = context.WithTimeout(context.Background(), timeout)
					t.Cleanup(cancel)
					wantErr = context.DeadlineExceeded
				case "operation-cancel":
					operation, cancel = context.WithCancel(context.Background())
					t.Cleanup(cancel)
				case "operation-already-canceled":
					cancelOperation()
					alreadyDone = true
				}

				started := make(chan string, 1)
				release := make(chan struct{})
				var requests atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					var req rpcRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						t.Error(err)
						return
					}
					if req.Method != method {
						t.Errorf("RPC method = %s, want %s", req.Method, method)
					}
					started <- r.Header.Get("X-Caller")
					if wantErr != nil && !alreadyDone {
						select {
						case <-r.Context().Done():
						case <-release:
						}
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x01"})
				}))
				t.Cleanup(func() {
					close(release)
					server.Close()
				})
				client, err := ethclient.Dial(server.URL)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(client.Close)
				backend := &operationBackend{Client: client, ctx: operation}
				done := make(chan error, 1)
				go func() {
					var result []byte
					var err error
					if method == "eth_call" {
						result, err = backend.CallContract(caller, ethereum.CallMsg{}, nil)
					} else {
						result, err = backend.CodeAt(caller, common.Address{}, nil)
					}
					if err == nil && string(result) != "\x01" {
						err = errors.New("unexpected successful RPC result")
					}
					done <- err
				}()
				if !alreadyDone {
					select {
					case header := <-started:
						if scenario == "caller-headers" && header != "preserved" {
							t.Fatalf("caller header = %q, want preserved", header)
						}
					case <-time.After(time.Second):
						t.Fatal("RPC did not reach local server")
					}
					if scenario == "caller-cancel" || scenario == "caller-cancel-with-deadline-cause" || scenario == "operation-cancel" {
						cancel()
					}
				}
				select {
				case err := <-done:
					if !errors.Is(err, wantErr) {
						t.Fatalf("RPC returned %v, want %v", err, wantErr)
					}
				case <-time.After(time.Second):
					t.Fatal("RPC ignored context cancellation or deadline")
				}
				if (scenario == "caller-deadline" || scenario == "caller-cancel" || scenario == "caller-cancel-with-deadline-cause") && operation.Err() != nil {
					t.Fatalf("caller cancellation ended the operation: %v", operation.Err())
				}
				if alreadyDone && requests.Load() != 0 {
					t.Fatal("RPC sent a request after operation had ended")
				}
			})
		}
	}
}
