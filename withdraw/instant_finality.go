package withdraw

import (
	"math/big"
	"strings"

	bindingspreview "github.com/ethereum-optimism/optimism/op-node/bindings/preview"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const proveAndFinalizeABI = `[{"inputs":[{"components":[{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"address","name":"sender","type":"address"},{"internalType":"address","name":"target","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"uint256","name":"gasLimit","type":"uint256"},{"internalType":"bytes","name":"data","type":"bytes"}],"internalType":"struct Types.WithdrawalTransaction","name":"_tx","type":"tuple"},{"internalType":"uint256","name":"_disputeGameIndex","type":"uint256"},{"components":[{"internalType":"bytes32","name":"version","type":"bytes32"},{"internalType":"bytes32","name":"stateRoot","type":"bytes32"},{"internalType":"bytes32","name":"messagePasserStorageRoot","type":"bytes32"},{"internalType":"bytes32","name":"latestBlockhash","type":"bytes32"}],"internalType":"struct Types.OutputRootProof","name":"_outputRootProof","type":"tuple"},{"internalType":"bytes[]","name":"_withdrawalProof","type":"bytes[]"}],"name":"proveAndFinalizeWithdrawalTransaction","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

// instantFinalityPortal provides access to the proveAndFinalizeWithdrawalTransaction
// method on OptimismPortal2 contracts configured for TEE-backed immediate finality.
type instantFinalityPortal struct {
	contract *bind.BoundContract
}

func newInstantFinalityPortal(address common.Address, backend bind.ContractBackend) (*instantFinalityPortal, error) {
	parsed, err := abi.JSON(strings.NewReader(proveAndFinalizeABI))
	if err != nil {
		return nil, err
	}
	contract := bind.NewBoundContract(address, parsed, backend, backend, backend)
	return &instantFinalityPortal{contract: contract}, nil
}

func (p *instantFinalityPortal) ProveAndFinalizeWithdrawalTransaction(
	opts *bind.TransactOpts,
	_tx bindingspreview.TypesWithdrawalTransaction,
	_disputeGameIndex *big.Int,
	_outputRootProof bindingspreview.TypesOutputRootProof,
	_withdrawalProof [][]byte,
) (*types.Transaction, error) {
	return p.contract.Transact(opts, "proveAndFinalizeWithdrawalTransaction", _tx, _disputeGameIndex, _outputRootProof, _withdrawalProof)
}
