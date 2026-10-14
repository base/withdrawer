package main

import (
	"fmt"
	"math/big"
	"net/http/httptest"
	"sync/atomic"
	"testing"

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
			helper, err := CreateWithdrawHelper(server.URL, common.Hash{}, network{
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
