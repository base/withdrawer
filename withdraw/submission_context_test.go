package withdraw

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestSubmissionHonorsBothContexts(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, action := range []string{"prove", "finalize"} {
			for _, source := range []string{"opts", "withdrawer", "withdrawer with nil opts context"} {
				for _, stage := range []string{"signing", "backoff", "expired deadline", "deadline during backoff"} {
					t.Run(fmt.Sprintf("fp=%t/%s/%s/%s", fp, action, source, stage), func(t *testing.T) {
						t.Parallel()
						ctx, cancel := context.WithCancel(context.Background())
						if stage == "expired deadline" {
							cancel()
							ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
						}
						defer cancel()
						var cancelDeadline context.CancelFunc
						defer func() {
							if cancelDeadline != nil {
								cancelDeadline()
							}
						}()
						var mu sync.Mutex
						sends, reconciliations := 0, 0
						w := rpcPathWithdrawer(t, fp, false, false, 1, func(req rpcRequest) {
							mu.Lock()
							defer mu.Unlock()
							if req.Method == "eth_sendRawTransaction" {
								sends++
								if stage == "backoff" {
									time.AfterFunc(10*time.Millisecond, cancel)
								}
							}
							if req.Method == "eth_getTransactionByHash" || req.Method == "eth_getTransactionReceipt" && string(req.Params[0]) != `"`+(common.Hash{}).Hex()+`"` {
								reconciliations++
							}
						})
						var opts *bind.TransactOpts
						var retry *RetryConfig
						switch w := w.(type) {
						case *Withdrawer:
							opts, retry = w.Opts, w.Retry
							if source != "opts" {
								w.Ctx = ctx
							}
						case *FPWithdrawer:
							opts, retry = w.Opts, w.Retry
							if source != "opts" {
								w.Ctx = ctx
							}
						}
						if source == "opts" {
							opts.Context = ctx
						} else if source == "withdrawer with nil opts context" {
							opts.Context = nil
						}
						signer := opts.Signer
						signed := 0
						opts.Signer = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
							signed++
							if stage == "signing" {
								cancel()
							} else if stage == "backoff" {
								retry.Delay = time.Second
							} else if stage == "deadline during backoff" {
								deadlineCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
								cancelDeadline = cancel
								if source == "opts" {
									opts.Context = deadlineCtx
								} else {
									switch w := w.(type) {
									case *Withdrawer:
										w.Ctx = deadlineCtx
									case *FPWithdrawer:
										w.Ctx = deadlineCtx
									}
								}
								retry.Delay = time.Second
							}
							return signer(address, tx)
						}
						var err error
						if action == "prove" {
							err = w.ProveWithdrawal()
						} else {
							err = w.FinalizeWithdrawal()
						}
						wantErr := context.Canceled
						if stage == "expired deadline" || stage == "deadline during backoff" {
							wantErr = context.DeadlineExceeded
						}
						mu.Lock()
						defer mu.Unlock()
						wantSends := 0
						if stage == "backoff" || stage == "deadline during backoff" {
							wantSends = 1
						}
						if !errors.Is(err, wantErr) || sends != wantSends || reconciliations != 0 || stage != "expired deadline" && signed != 1 {
							t.Fatalf("error %v, signed %d, sends %d, reconciliation queries %d", err, signed, sends, reconciliations)
						}
					})
				}
			}
		}
	}
}

func TestSubmissionWithNilOptsContext(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, action := range []string{"prove", "finalize"} {
			t.Run(fmt.Sprintf("fp=%t/%s", fp, action), func(t *testing.T) {
				t.Parallel()
				w := rpcPathWithdrawer(t, fp, false, false, 1)
				switch w := w.(type) {
				case *Withdrawer:
					w.Opts.Context = nil
				case *FPWithdrawer:
					w.Opts.Context = nil
				}
				var err error
				if action == "prove" {
					err = w.ProveWithdrawal()
				} else {
					err = w.FinalizeWithdrawal()
				}
				if err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestSubmissionPreservesOptsContextHeaders(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, action := range []string{"prove", "finalize"} {
			t.Run(fmt.Sprintf("fp=%t/%s", fp, action), func(t *testing.T) {
				t.Parallel()
				var mu sync.Mutex
				preparations, sends := 0, 0
				var payload string
				w := rpcPathWithdrawer(t, fp, false, false, 1, func(req rpcRequest) {
					mu.Lock()
					defer mu.Unlock()
					switch req.Method {
					case "eth_maxPriorityFeePerGas", "eth_estimateGas", "eth_getTransactionCount":
						preparations++
					case "eth_sendRawTransaction":
						sends++
						if sends == 1 {
							payload = string(req.Params[0])
						} else if string(req.Params[0]) != payload {
							t.Error("retry changed the signed transaction")
						}
					default:
						return
					}
					if req.Header.Get("Authorization") != "Bearer transaction" || req.Header.Get("X-Rpc-Route") != "transaction" {
						t.Errorf("%s lost transaction context headers: %v", req.Method, req.Header)
					}
				})
				var opts *bind.TransactOpts
				withdrawerHeaders := http.Header{"Authorization": {"Bearer withdrawer"}, "X-Rpc-Route": {"withdrawer"}}
				switch w := w.(type) {
				case *Withdrawer:
					w.Ctx = rpc.NewContextWithHeaders(w.Ctx, withdrawerHeaders)
					opts = w.Opts
				case *FPWithdrawer:
					w.Ctx = rpc.NewContextWithHeaders(w.Ctx, withdrawerHeaders)
					opts = w.Opts
				}
				opts.Context = rpc.NewContextWithHeaders(opts.Context, http.Header{"Authorization": {"Bearer transaction"}, "X-Rpc-Route": {"transaction"}})
				signer := opts.Signer
				signed := 0
				opts.Signer = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
					signed++
					return signer(address, tx)
				}
				var err error
				if action == "prove" {
					err = w.ProveWithdrawal()
				} else {
					err = w.FinalizeWithdrawal()
				}
				if err != nil {
					t.Fatal(err)
				}
				mu.Lock()
				defer mu.Unlock()
				if preparations == 0 || sends != 2 || signed != 1 {
					t.Fatalf("preparation calls %d, sends %d, signatures %d", preparations, sends, signed)
				}
			})
		}
	}
}
