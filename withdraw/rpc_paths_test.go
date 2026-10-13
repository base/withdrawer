package withdraw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"testing"
	"time"

	withdrawerbindings "github.com/base/withdrawer/bindings"
	"github.com/ethereum-optimism/optimism/op-node/bindings"
	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

func trieProof(t *testing.T, tree *trie.Trie, key []byte) []string {
	t.Helper()
	db := memorydb.New()
	defer db.Close()
	if err := tree.Prove(key, db); err != nil {
		t.Fatal(err)
	}
	it := db.NewIterator(nil, nil)
	defer it.Release()
	var nodes []string
	for it.Next() {
		nodes = append(nodes, hexutil.Encode(it.Value()))
	}
	if err := it.Error(); err != nil {
		t.Fatal(err)
	}
	return nodes
}

func withdrawalProof(t *testing.T, receipt *types.Receipt) (common.Hash, any) {
	t.Helper()
	ev, err := withdrawals.ParseMessagePassed(receipt)
	if err != nil {
		t.Fatal(err)
	}
	slot := withdrawals.StorageSlotOfWithdrawalHash(ev.WithdrawalHash)
	storage := trie.NewEmpty(nil)
	storageKey := crypto.Keccak256(slot[:])
	if err := storage.Update(storageKey, []byte{1}); err != nil {
		t.Fatal(err)
	}
	account := types.NewEmptyStateAccount()
	account.Root = storage.Hash()
	encoded, err := rlp.EncodeToBytes(account)
	if err != nil {
		t.Fatal(err)
	}
	state := trie.NewEmpty(nil)
	address := predeploys.L2ToL1MessagePasserAddr
	accountKey := crypto.Keccak256(address[:])
	if err := state.Update(accountKey, encoded); err != nil {
		t.Fatal(err)
	}
	return state.Hash(), map[string]any{
		"address": address, "balance": "0x0", "nonce": "0x0", "codeHash": types.EmptyCodeHash,
		"storageHash": account.Root, "accountProof": trieProof(t, state, accountKey),
		"storageProof": []any{map[string]any{"key": slot.Hex(), "value": "0x1", "proof": trieProof(t, storage, storageKey)}},
	}
}

func rpcPathWithdrawer(t *testing.T, fp, dry bool, multiplier float64) WithdrawHelper {
	t.Helper()
	receipt := messageReceipt(t)
	root, proof := withdrawalProof(t, receipt)
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	opts, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	opts.Context = context.Background()
	methods := map[string]abi.Method{}
	for _, metadata := range []*bind.MetaData{bindings.OptimismPortalMetaData, bindings.L2OutputOracleMetaData, bindingspreview.OptimismPortal2MetaData, bindings.DisputeGameFactoryMetaData, withdrawerbindings.AnchorStateRegistryMetaData} {
		parsed, err := metadata.GetAbi()
		if err != nil {
			t.Fatal(err)
		}
		for _, method := range parsed.Methods {
			methods[hexutil.Encode(method.ID)] = method
		}
	}
	var mu sync.Mutex
	failed := map[string]bool{}
	var payloads []string
	var sentHash common.Hash
	client := localRPC(t, func(req rpcRequest) (any, *rpcFailure, int) {
		mu.Lock()
		defer mu.Unlock()
		if len(payloads) > 0 && req.Method == "eth_getTransactionCount" {
			return nil, &rpcFailure{-32000, "nonce queried after submission"}, 0
		}
		faultKey := req.Method
		if len(req.Params) > 0 {
			faultKey += string(req.Params[0])
		}
		if !failed[faultKey] {
			failed[faultKey] = true
			return nil, nil, http.StatusServiceUnavailable
		}
		switch req.Method {
		case "eth_getTransactionReceipt":
			var hash common.Hash
			if err := json.Unmarshal(req.Params[0], &hash); err != nil {
				t.Error(err)
			}
			if hash == sentHash && sentHash != (common.Hash{}) {
				return successfulReceipt(), nil, 0
			}
			return receipt, nil, 0
		case "eth_getProof":
			return proof, nil, 0
		case "eth_getBlockByNumber":
			var number string
			if err := json.Unmarshal(req.Params[0], &number); err != nil {
				t.Error(err)
			}
			timestamp := uint64(10)
			if number == "latest" {
				timestamp = 1000
			}
			return &types.Header{Number: big.NewInt(5), Difficulty: big.NewInt(0), Root: root, Time: timestamp, BaseFee: big.NewInt(10)}, nil, 0
		case "eth_getCode":
			return "0x01", nil, 0
		case "eth_maxPriorityFeePerGas":
			return "0x1", nil, 0
		case "eth_estimateGas":
			return "0x10000", nil, 0
		case "eth_getTransactionCount":
			return "0x7", nil, 0
		case "eth_call":
			var msg struct {
				Input hexutil.Bytes `json:"input"`
			}
			if err := json.Unmarshal(req.Params[0], &msg); err != nil {
				t.Error(err)
			}
			if len(msg.Input) < 4 {
				t.Error("missing selector")
				return nil, &rpcFailure{-32602, "missing input"}, 0
			}
			method, ok := methods[hexutil.Encode(msg.Input[:4])]
			if !ok {
				t.Errorf("unknown selector %x", msg.Input[:4])
				return nil, &rpcFailure{-32601, "unknown method"}, 0
			}
			var values []any
			switch method.RawName {
			case "SUBMISSION_INTERVAL", "L2_BLOCK_TIME", "FINALIZATION_PERIOD_SECONDS":
				values = []any{big.NewInt(1)}
			case "latestBlockNumber":
				values = []any{big.NewInt(5)}
			case "getL2OutputIndexAfter":
				values = []any{big.NewInt(0)}
			case "provenWithdrawals":
				if len(method.Inputs) == 1 {
					values = []any{[32]byte{}, big.NewInt(123), big.NewInt(0)}
				} else {
					values = []any{common.HexToAddress("0x08"), uint64(123)}
				}
			case "respectedGameType":
				values = []any{uint32(0)}
			case "gameCount":
				values = []any{big.NewInt(1)}
			case "findLatestGames":
				values = []any{[]bindings.IDisputeGameFactoryGameSearchResult{{Index: big.NewInt(0), ExtraData: common.LeftPadBytes([]byte{5}, 32)}}}
			case "isGameBlacklisted", "isGameRetired", "finalizedWithdrawals":
				values = []any{false}
			case "checkWithdrawal":
			default:
				t.Errorf("unexpected contract read %s", method.RawName)
				return nil, &rpcFailure{-32601, "unexpected method"}, 0
			}
			data, err := method.Outputs.Pack(values...)
			if err != nil {
				t.Error(err)
				return nil, &rpcFailure{-32603, err.Error()}, 0
			}
			return hexutil.Encode(data), nil, 0
		case "eth_sendRawTransaction":
			var raw string
			if err := json.Unmarshal(req.Params[0], &raw); err != nil {
				t.Error(err)
			}
			payloads = append(payloads, raw)
			var tx types.Transaction
			data, err := hexutil.Decode(raw)
			if err != nil {
				t.Error(err)
			}
			if err := tx.UnmarshalBinary(data); err != nil {
				t.Error(err)
			}
			sentHash = tx.Hash()
			if tx.Gas() != uint64(65536*multiplier) || tx.Nonce() != 7 {
				t.Errorf("signed gas %d nonce %d", tx.Gas(), tx.Nonce())
			}
			return nil, &rpcFailure{-32000, "already known"}, 0
		default:
			t.Errorf("unexpected RPC %s", req.Method)
			return nil, &rpcFailure{-32601, "unknown method"}, 0
		}
	})
	config := &RetryConfig{MaxRetries: 1, Delay: time.Millisecond}
	l1 := ethclient.NewClient(client)
	backend := NewRetryingBackend(opts.Context, l1, config)
	portalAddress := common.HexToAddress("0x01")
	oracleAddress := common.HexToAddress("0x02")
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		if dry && len(payloads) > 0 {
			t.Errorf("dry-run submitted a transaction")
		}
	})
	if fp {
		portal, err := bindingspreview.NewOptimismPortal2(portalAddress, backend)
		if err != nil {
			t.Fatal(err)
		}
		factory, err := bindings.NewDisputeGameFactory(oracleAddress, backend)
		if err != nil {
			t.Fatal(err)
		}
		registry, err := withdrawerbindings.NewAnchorStateRegistry(common.HexToAddress("0x03"), backend)
		if err != nil {
			t.Fatal(err)
		}
		return &FPWithdrawer{Ctx: opts.Context, L1Client: l1, L2Client: client, Portal: portal, Factory: factory, AnchorStateRegistry: registry, Opts: opts, GasMultiplier: multiplier, DryRun: dry, Retry: config}
	}
	portal, err := bindings.NewOptimismPortal(portalAddress, backend)
	if err != nil {
		t.Fatal(err)
	}
	oracle, err := bindings.NewL2OutputOracle(oracleAddress, backend)
	if err != nil {
		t.Fatal(err)
	}
	return &Withdrawer{Ctx: opts.Context, L1Client: l1, L2Client: client, Portal: portal, Oracle: oracle, Opts: opts, GasMultiplier: multiplier, DryRun: dry, Retry: config}
}

func TestWithdrawalRPCPathsRecover(t *testing.T) {
	for _, fp := range []bool{false, true} {
		for _, action := range []string{"prove", "finalize"} {
			for _, dry := range []bool{false, true} {
				t.Run(fmt.Sprintf("fp=%t/%s/dry=%t", fp, action, dry), func(t *testing.T) {
					t.Parallel()
					w := rpcPathWithdrawer(t, fp, dry, 1.5)
					if err := w.CheckIfProvable(); err != nil {
						t.Fatal(err)
					}
					proven, err := w.GetProvenWithdrawalTime()
					if err != nil || proven != 123 {
						t.Fatalf("proven time %d: %v", proven, err)
					}
					finalized, err := w.IsProofFinalized()
					if err != nil || finalized {
						t.Fatalf("finalized %t: %v", finalized, err)
					}
					if action == "prove" {
						err = w.ProveWithdrawal()
					} else {
						err = w.FinalizeWithdrawal()
					}
					if err != nil {
						t.Fatalf("%s did not recover: %v", action, err)
					}
				})
			}
		}
	}
}

func TestSignerErrorsAreNotRetried(t *testing.T) {
	w := rpcPathWithdrawer(t, true, false, 1).(*FPWithdrawer)
	signer := w.Opts.Signer
	rejected := false
	w.Opts.Signer = func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if !rejected {
			rejected = true
			return nil, rpc.HTTPError{StatusCode: 503, Status: "signer unavailable"}
		}
		return signer(address, tx)
	}
	err := w.FinalizeWithdrawal()
	var signerErr rpc.HTTPError
	if !errors.As(err, &signerErr) || signerErr.Status != "signer unavailable" {
		t.Fatalf("signer failure was retried or hidden: %v", err)
	}
}
