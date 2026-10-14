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

// A stable Done channel lets submission subscribe before arming and publishes timeout to Err.
type submissionDeadlineContext struct {
	context.Context
	done    chan struct{}
	timeout context.Context
}

func (c *submissionDeadlineContext) Done() <-chan struct{} { return c.done }

func (c *submissionDeadlineContext) Err() error {
	select {
	case <-c.done:
		return c.timeout.Err()
	default:
		return nil
	}
}

func (c *submissionDeadlineContext) arm(delay time.Duration) context.CancelFunc {
	var cancel context.CancelFunc
	c.timeout, cancel = context.WithTimeout(c.Context, delay)
	context.AfterFunc(c.timeout, func() { close(c.done) })
	return cancel
}

func TestSubmissionHonorsBothContexts(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, action := range []string{"prove", "finalize"} {
			for _, source := range []string{"opts", "withdrawer", "withdrawer with nil opts context"} {
				for _, stage := range []string{"signing", "backoff", "expired deadline", "deadline during backoff", "deadline during backoff after delayed signing"} {
					t.Run(fmt.Sprintf("fp=%t/%s/%s/%s", fp, action, source, stage), func(t *testing.T) {
						t.Parallel()
						ctx, cancel := context.WithCancel(context.Background())
						if stage == "expired deadline" {
							cancel()
							ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
						}
						defer cancel()
						deadlineDuringBackoff := stage == "deadline during backoff" || stage == "deadline during backoff after delayed signing"
						var deadline *submissionDeadlineContext
						if deadlineDuringBackoff {
							deadline = &submissionDeadlineContext{Context: ctx, done: make(chan struct{})}
							ctx = deadline
						}
						var mu sync.Mutex
						sends, reconciliations := 0, 0
						w := rpcPathWithdrawer(t, fp, false, false, 1, func(req rpcRequest) {
							mu.Lock()
							defer mu.Unlock()
							if req.Method == "eth_sendRawTransaction" {
								sends++
								if deadlineDuringBackoff && sends == 1 {
									t.Cleanup(deadline.arm(200 * time.Millisecond))
								}
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
							} else if stage == "backoff" || deadlineDuringBackoff {
								retry.Delay = time.Second
							}
							signedTx, err := signer(address, tx)
							if stage == "deadline during backoff after delayed signing" {
								time.Sleep(300 * time.Millisecond)
							}
							return signedTx, err
						}
						var err error
						if action == "prove" {
							err = w.ProveWithdrawal()
						} else {
							err = w.FinalizeWithdrawal()
						}
						wantErr := context.Canceled
						if stage == "expired deadline" || deadlineDuringBackoff {
							wantErr = context.DeadlineExceeded
						}
						mu.Lock()
						defer mu.Unlock()
						wantSends := 0
						if stage == "backoff" || deadlineDuringBackoff {
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
