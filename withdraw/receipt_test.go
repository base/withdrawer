package withdraw

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type receiptService struct {
	receipt func() (*types.Receipt, error)
}

func (s *receiptService) GetTransactionReceipt(common.Hash) (*types.Receipt, error) {
	return s.receipt()
}

func (s *receiptService) GetBlockByNumber(string, bool) (*types.Header, error) {
	return nil, receiptRPCError{}
}

type receiptRPCError struct{}

func (receiptRPCError) Error() string  { return "receipt RPC unavailable" }
func (receiptRPCError) ErrorCode() int { return -32042 }

func receiptClient(t *testing.T, fetch func() (*types.Receipt, error)) *rpc.Client {
	t.Helper()
	server := rpc.NewServer()
	if err := server.RegisterName("eth", &receiptService{receipt: fetch}); err != nil {
		t.Fatal(err)
	}
	client := rpc.DialInProc(server)
	t.Cleanup(client.Close)
	t.Cleanup(server.Stop)
	return client
}

func receiptFixture(status uint64) *types.Receipt {
	return &types.Receipt{
		Status:            status,
		TxHash:            common.HexToHash("0xdead"),
		BlockNumber:       big.NewInt(42),
		CumulativeGasUsed: 21000,
		Logs:              []*types.Log{},
	}
}

func requireErrorContext(t *testing.T, err error, parts ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, part := range parts {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error %q is missing %q", err, part)
		}
	}
}

var l2ReceiptPaths = []struct {
	name string
	run  func(context.Context, *rpc.Client, common.Hash) error
}{
	{"txBlock", func(ctx context.Context, client *rpc.Client, hash common.Hash) error {
		_, err := txBlock(ctx, client, hash)
		return err
	}},
	{"FinalizeWithdrawal", func(ctx context.Context, client *rpc.Client, hash common.Hash) error {
		w := &Withdrawer{Ctx: ctx, L2Client: client, L2TxHash: hash}
		return w.FinalizeWithdrawal()
	}},
}

func TestL2WithdrawalReceiptFailure(t *testing.T) {
	hash := common.HexToHash("0x1234")
	for _, path := range l2ReceiptPaths {
		for _, status := range []uint64{types.ReceiptStatusFailed, 2} {
			t.Run(fmt.Sprintf("%s/status_%d", path.name, status), func(t *testing.T) {
				client := receiptClient(t, func() (*types.Receipt, error) {
					return receiptFixture(status), nil
				})
				err := path.run(context.Background(), client, hash)
				requireErrorContext(t, err, "L2 withdrawal transaction failed", hash.Hex(), fmt.Sprintf("status %d", status))
			})
		}
	}
}

func TestL2WithdrawalReceiptFetchErrors(t *testing.T) {
	hash := common.HexToHash("0x1234")
	for _, path := range l2ReceiptPaths {
		for _, failure := range []string{"not_found", "rpc", "canceled"} {
			t.Run(path.name+"/"+failure, func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				if failure == "canceled" {
					cancel()
				}
				client := receiptClient(t, func() (*types.Receipt, error) {
					if failure == "rpc" {
						return nil, receiptRPCError{}
					}
					return nil, nil
				})
				err := path.run(ctx, client, hash)
				requireErrorContext(t, err, "L2 withdrawal transaction", hash.Hex())
				switch failure {
				case "not_found":
					if !errors.Is(err, ethereum.NotFound) {
						t.Errorf("expected wrapped NotFound, got %v", err)
					}
				case "canceled":
					if !errors.Is(err, context.Canceled) {
						t.Errorf("expected wrapped cancellation, got %v", err)
					}
				case "rpc":
					var rpcErr rpc.Error
					if !errors.As(err, &rpcErr) || rpcErr.ErrorCode() != -32042 {
						t.Errorf("expected wrapped RPC error -32042, got %v", err)
					}
				}
			})
		}
	}
}

func TestL2WithdrawalReceiptSuccess(t *testing.T) {
	client := receiptClient(t, func() (*types.Receipt, error) {
		return receiptFixture(types.ReceiptStatusSuccessful), nil
	})
	hash := common.HexToHash("0x1234")
	block, err := txBlock(context.Background(), client, hash)
	if err != nil || block == nil || block.Cmp(big.NewInt(42)) != 0 {
		t.Fatalf("expected withdrawal block 42, got %v, %v", block, err)
	}
	w := &Withdrawer{Ctx: context.Background(), L2Client: client, L2TxHash: hash}
	requireErrorContext(t, w.FinalizeWithdrawal(), "error getting header by number for block 42")
}

func TestL1Confirmation(t *testing.T) {
	hash := common.HexToHash("0x5678")
	for _, operation := range []string{"proof", "finalization"} {
		for _, outcome := range []string{"success", "status_0", "status_2", "rpc", "canceled", "polling_deadline"} {
			t.Run(operation+"/"+outcome, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if outcome == "canceled" {
					cancel()
				}
				client := receiptClient(t, func() (*types.Receipt, error) {
					switch outcome {
					case "status_0":
						return receiptFixture(types.ReceiptStatusFailed), nil
					case "status_2":
						return receiptFixture(2), nil
					case "rpc":
						return nil, receiptRPCError{}
					case "polling_deadline":
						return nil, nil
					default:
						return receiptFixture(types.ReceiptStatusSuccessful), nil
					}
				})
				err := waitForConfirmation(ctx, ethclient.NewClient(client), hash, operation)
				if outcome == "success" {
					if err != nil {
						t.Fatalf("successful receipt rejected: %v", err)
					}
					return
				}
				requireErrorContext(t, err, "L1 "+operation+" transaction", hash.Hex())
				switch outcome {
				case "status_0", "status_2":
					requireErrorContext(t, err, "transaction failed", "status "+strings.TrimPrefix(outcome, "status_"))
				case "rpc":
					var rpcErr rpc.Error
					if !errors.As(err, &rpcErr) || rpcErr.ErrorCode() != -32042 {
						t.Errorf("expected wrapped RPC error -32042, got %v", err)
					}
				case "canceled":
					if !errors.Is(err, context.Canceled) {
						t.Errorf("expected wrapped cancellation, got %v", err)
					}
				case "polling_deadline":
					if !errors.Is(err, context.DeadlineExceeded) {
						t.Errorf("expected wrapped polling deadline, got %v", err)
					}
				}
			})
		}
	}
}

func TestL1ConfirmationPollsNotFound(t *testing.T) {
	attempts := 0
	client := receiptClient(t, func() (*types.Receipt, error) {
		attempts++
		if attempts == 1 {
			return nil, nil
		}
		return receiptFixture(types.ReceiptStatusSuccessful), nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := waitForConfirmation(ctx, ethclient.NewClient(client), common.HexToHash("0x5678"), "proof"); err != nil {
		t.Fatalf("receipt did not confirm after NotFound: %v", err)
	}
}
