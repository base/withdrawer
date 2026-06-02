// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// Proposal is an auto generated low-level Go binding around an user-defined struct.
type Proposal struct {
	Root             [32]byte
	L2SequenceNumber *big.Int
}

// AnchorStateRegistryMetaData contains all meta data concerning the AnchorStateRegistry contract.
var AnchorStateRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_disputeGameFinalityDelaySeconds\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"anchorGame\",\"outputs\":[{\"internalType\":\"contractIFaultDisputeGame\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"GameType\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"anchors\",\"outputs\":[{\"internalType\":\"Hash\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_disputeGame\",\"type\":\"address\"}],\"name\":\"blacklistDisputeGame\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"disputeGameBlacklist\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"disputeGameFactory\",\"outputs\":[{\"internalType\":\"contractIDisputeGameFactory\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"disputeGameFinalityDelaySeconds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAnchorRoot\",\"outputs\":[{\"internalType\":\"Hash\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getStartingAnchorRoot\",\"outputs\":[{\"components\":[{\"internalType\":\"Hash\",\"name\":\"root\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"l2SequenceNumber\",\"type\":\"uint256\"}],\"internalType\":\"structProposal\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initVersion\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractISystemConfig\",\"name\":\"_systemConfig\",\"type\":\"address\"},{\"internalType\":\"contractIDisputeGameFactory\",\"name\":\"_disputeGameFactory\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"Hash\",\"name\":\"root\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"l2SequenceNumber\",\"type\":\"uint256\"}],\"internalType\":\"structProposal\",\"name\":\"_startingAnchorRoot\",\"type\":\"tuple\"},{\"internalType\":\"GameType\",\"name\":\"_startingRespectedGameType\",\"type\":\"uint32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameBlacklisted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameClaimValid\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameFinalized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameProper\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameRegistered\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameResolved\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameRespected\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"isGameRetired\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxyAdmin\",\"outputs\":[{\"internalType\":\"contractIProxyAdmin\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"proxyAdminOwner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"respectedGameType\",\"outputs\":[{\"internalType\":\"GameType\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"retirementTimestamp\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIDisputeGame\",\"name\":\"_game\",\"type\":\"address\"}],\"name\":\"setAnchorState\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"GameType\",\"name\":\"_gameType\",\"type\":\"uint32\"}],\"name\":\"setRespectedGameType\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"superchainConfig\",\"outputs\":[{\"internalType\":\"contractISuperchainConfig\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"systemConfig\",\"outputs\":[{\"internalType\":\"contractISystemConfig\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"updateRetirementTimestamp\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIFaultDisputeGame\",\"name\":\"game\",\"type\":\"address\"}],\"name\":\"AnchorUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"contractIDisputeGame\",\"name\":\"disputeGame\",\"type\":\"address\"}],\"name\":\"DisputeGameBlacklisted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"GameType\",\"name\":\"gameType\",\"type\":\"uint32\"}],\"name\":\"RespectedGameTypeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"RetirementTimestampSet\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"AnchorStateRegistry_InvalidAnchorGame\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"AnchorStateRegistry_Unauthorized\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdmin\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdminOrProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotResolvedDelegateProxy\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_NotSharedProxyAdminOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProxyAdminOwnedBase_ProxyAdminNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ReinitializableBase_ZeroInitVersion\",\"type\":\"error\"}]",
}

// AnchorStateRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AnchorStateRegistryMetaData.ABI instead.
var AnchorStateRegistryABI = AnchorStateRegistryMetaData.ABI

// AnchorStateRegistry is an auto generated Go binding around an Ethereum contract.
type AnchorStateRegistry struct {
	AnchorStateRegistryCaller     // Read-only binding to the contract
	AnchorStateRegistryTransactor // Write-only binding to the contract
	AnchorStateRegistryFilterer   // Log filterer for contract events
}

// AnchorStateRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AnchorStateRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorStateRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AnchorStateRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorStateRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AnchorStateRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AnchorStateRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AnchorStateRegistrySession struct {
	Contract     *AnchorStateRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// AnchorStateRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AnchorStateRegistryCallerSession struct {
	Contract *AnchorStateRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// AnchorStateRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AnchorStateRegistryTransactorSession struct {
	Contract     *AnchorStateRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// AnchorStateRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AnchorStateRegistryRaw struct {
	Contract *AnchorStateRegistry // Generic contract binding to access the raw methods on
}

// AnchorStateRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AnchorStateRegistryCallerRaw struct {
	Contract *AnchorStateRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// AnchorStateRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AnchorStateRegistryTransactorRaw struct {
	Contract *AnchorStateRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAnchorStateRegistry creates a new instance of AnchorStateRegistry, bound to a specific deployed contract.
func NewAnchorStateRegistry(address common.Address, backend bind.ContractBackend) (*AnchorStateRegistry, error) {
	contract, err := bindAnchorStateRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistry{AnchorStateRegistryCaller: AnchorStateRegistryCaller{contract: contract}, AnchorStateRegistryTransactor: AnchorStateRegistryTransactor{contract: contract}, AnchorStateRegistryFilterer: AnchorStateRegistryFilterer{contract: contract}}, nil
}

// NewAnchorStateRegistryCaller creates a new read-only instance of AnchorStateRegistry, bound to a specific deployed contract.
func NewAnchorStateRegistryCaller(address common.Address, caller bind.ContractCaller) (*AnchorStateRegistryCaller, error) {
	contract, err := bindAnchorStateRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryCaller{contract: contract}, nil
}

// NewAnchorStateRegistryTransactor creates a new write-only instance of AnchorStateRegistry, bound to a specific deployed contract.
func NewAnchorStateRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AnchorStateRegistryTransactor, error) {
	contract, err := bindAnchorStateRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryTransactor{contract: contract}, nil
}

// NewAnchorStateRegistryFilterer creates a new log filterer instance of AnchorStateRegistry, bound to a specific deployed contract.
func NewAnchorStateRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AnchorStateRegistryFilterer, error) {
	contract, err := bindAnchorStateRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryFilterer{contract: contract}, nil
}

// bindAnchorStateRegistry binds a generic wrapper to an already deployed contract.
func bindAnchorStateRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AnchorStateRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnchorStateRegistry *AnchorStateRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnchorStateRegistry.Contract.AnchorStateRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnchorStateRegistry *AnchorStateRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.AnchorStateRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnchorStateRegistry *AnchorStateRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.AnchorStateRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AnchorStateRegistry *AnchorStateRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AnchorStateRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AnchorStateRegistry *AnchorStateRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AnchorStateRegistry *AnchorStateRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.contract.Transact(opts, method, params...)
}

// AnchorGame is a free data retrieval call binding the contract method 0xe0a840eb.
//
// Solidity: function anchorGame() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) AnchorGame(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "anchorGame")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AnchorGame is a free data retrieval call binding the contract method 0xe0a840eb.
//
// Solidity: function anchorGame() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) AnchorGame() (common.Address, error) {
	return _AnchorStateRegistry.Contract.AnchorGame(&_AnchorStateRegistry.CallOpts)
}

// AnchorGame is a free data retrieval call binding the contract method 0xe0a840eb.
//
// Solidity: function anchorGame() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) AnchorGame() (common.Address, error) {
	return _AnchorStateRegistry.Contract.AnchorGame(&_AnchorStateRegistry.CallOpts)
}

// Anchors is a free data retrieval call binding the contract method 0x7258a807.
//
// Solidity: function anchors(uint32 ) view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) Anchors(opts *bind.CallOpts, arg0 uint32) ([32]byte, *big.Int, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "anchors", arg0)

	if err != nil {
		return *new([32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// Anchors is a free data retrieval call binding the contract method 0x7258a807.
//
// Solidity: function anchors(uint32 ) view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistrySession) Anchors(arg0 uint32) ([32]byte, *big.Int, error) {
	return _AnchorStateRegistry.Contract.Anchors(&_AnchorStateRegistry.CallOpts, arg0)
}

// Anchors is a free data retrieval call binding the contract method 0x7258a807.
//
// Solidity: function anchors(uint32 ) view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) Anchors(arg0 uint32) ([32]byte, *big.Int, error) {
	return _AnchorStateRegistry.Contract.Anchors(&_AnchorStateRegistry.CallOpts, arg0)
}

// DisputeGameBlacklist is a free data retrieval call binding the contract method 0x45884d32.
//
// Solidity: function disputeGameBlacklist(address ) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) DisputeGameBlacklist(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "disputeGameBlacklist", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// DisputeGameBlacklist is a free data retrieval call binding the contract method 0x45884d32.
//
// Solidity: function disputeGameBlacklist(address ) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) DisputeGameBlacklist(arg0 common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.DisputeGameBlacklist(&_AnchorStateRegistry.CallOpts, arg0)
}

// DisputeGameBlacklist is a free data retrieval call binding the contract method 0x45884d32.
//
// Solidity: function disputeGameBlacklist(address ) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) DisputeGameBlacklist(arg0 common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.DisputeGameBlacklist(&_AnchorStateRegistry.CallOpts, arg0)
}

// DisputeGameFactory is a free data retrieval call binding the contract method 0xf2b4e617.
//
// Solidity: function disputeGameFactory() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) DisputeGameFactory(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "disputeGameFactory")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DisputeGameFactory is a free data retrieval call binding the contract method 0xf2b4e617.
//
// Solidity: function disputeGameFactory() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) DisputeGameFactory() (common.Address, error) {
	return _AnchorStateRegistry.Contract.DisputeGameFactory(&_AnchorStateRegistry.CallOpts)
}

// DisputeGameFactory is a free data retrieval call binding the contract method 0xf2b4e617.
//
// Solidity: function disputeGameFactory() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) DisputeGameFactory() (common.Address, error) {
	return _AnchorStateRegistry.Contract.DisputeGameFactory(&_AnchorStateRegistry.CallOpts)
}

// DisputeGameFinalityDelaySeconds is a free data retrieval call binding the contract method 0x952b2797.
//
// Solidity: function disputeGameFinalityDelaySeconds() view returns(uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) DisputeGameFinalityDelaySeconds(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "disputeGameFinalityDelaySeconds")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DisputeGameFinalityDelaySeconds is a free data retrieval call binding the contract method 0x952b2797.
//
// Solidity: function disputeGameFinalityDelaySeconds() view returns(uint256)
func (_AnchorStateRegistry *AnchorStateRegistrySession) DisputeGameFinalityDelaySeconds() (*big.Int, error) {
	return _AnchorStateRegistry.Contract.DisputeGameFinalityDelaySeconds(&_AnchorStateRegistry.CallOpts)
}

// DisputeGameFinalityDelaySeconds is a free data retrieval call binding the contract method 0x952b2797.
//
// Solidity: function disputeGameFinalityDelaySeconds() view returns(uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) DisputeGameFinalityDelaySeconds() (*big.Int, error) {
	return _AnchorStateRegistry.Contract.DisputeGameFinalityDelaySeconds(&_AnchorStateRegistry.CallOpts)
}

// GetAnchorRoot is a free data retrieval call binding the contract method 0xd83ef267.
//
// Solidity: function getAnchorRoot() view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) GetAnchorRoot(opts *bind.CallOpts) ([32]byte, *big.Int, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "getAnchorRoot")

	if err != nil {
		return *new([32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetAnchorRoot is a free data retrieval call binding the contract method 0xd83ef267.
//
// Solidity: function getAnchorRoot() view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistrySession) GetAnchorRoot() ([32]byte, *big.Int, error) {
	return _AnchorStateRegistry.Contract.GetAnchorRoot(&_AnchorStateRegistry.CallOpts)
}

// GetAnchorRoot is a free data retrieval call binding the contract method 0xd83ef267.
//
// Solidity: function getAnchorRoot() view returns(bytes32, uint256)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) GetAnchorRoot() ([32]byte, *big.Int, error) {
	return _AnchorStateRegistry.Contract.GetAnchorRoot(&_AnchorStateRegistry.CallOpts)
}

// GetStartingAnchorRoot is a free data retrieval call binding the contract method 0x664ed830.
//
// Solidity: function getStartingAnchorRoot() view returns((bytes32,uint256))
func (_AnchorStateRegistry *AnchorStateRegistryCaller) GetStartingAnchorRoot(opts *bind.CallOpts) (Proposal, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "getStartingAnchorRoot")

	if err != nil {
		return *new(Proposal), err
	}

	out0 := *abi.ConvertType(out[0], new(Proposal)).(*Proposal)

	return out0, err

}

// GetStartingAnchorRoot is a free data retrieval call binding the contract method 0x664ed830.
//
// Solidity: function getStartingAnchorRoot() view returns((bytes32,uint256))
func (_AnchorStateRegistry *AnchorStateRegistrySession) GetStartingAnchorRoot() (Proposal, error) {
	return _AnchorStateRegistry.Contract.GetStartingAnchorRoot(&_AnchorStateRegistry.CallOpts)
}

// GetStartingAnchorRoot is a free data retrieval call binding the contract method 0x664ed830.
//
// Solidity: function getStartingAnchorRoot() view returns((bytes32,uint256))
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) GetStartingAnchorRoot() (Proposal, error) {
	return _AnchorStateRegistry.Contract.GetStartingAnchorRoot(&_AnchorStateRegistry.CallOpts)
}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) InitVersion(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "initVersion")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_AnchorStateRegistry *AnchorStateRegistrySession) InitVersion() (uint8, error) {
	return _AnchorStateRegistry.Contract.InitVersion(&_AnchorStateRegistry.CallOpts)
}

// InitVersion is a free data retrieval call binding the contract method 0x38d38c97.
//
// Solidity: function initVersion() view returns(uint8)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) InitVersion() (uint8, error) {
	return _AnchorStateRegistry.Contract.InitVersion(&_AnchorStateRegistry.CallOpts)
}

// IsGameBlacklisted is a free data retrieval call binding the contract method 0x34a346ea.
//
// Solidity: function isGameBlacklisted(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameBlacklisted(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameBlacklisted", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameBlacklisted is a free data retrieval call binding the contract method 0x34a346ea.
//
// Solidity: function isGameBlacklisted(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameBlacklisted(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameBlacklisted(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameBlacklisted is a free data retrieval call binding the contract method 0x34a346ea.
//
// Solidity: function isGameBlacklisted(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameBlacklisted(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameBlacklisted(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameClaimValid is a free data retrieval call binding the contract method 0x6c4f4467.
//
// Solidity: function isGameClaimValid(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameClaimValid(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameClaimValid", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameClaimValid is a free data retrieval call binding the contract method 0x6c4f4467.
//
// Solidity: function isGameClaimValid(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameClaimValid(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameClaimValid(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameClaimValid is a free data retrieval call binding the contract method 0x6c4f4467.
//
// Solidity: function isGameClaimValid(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameClaimValid(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameClaimValid(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameFinalized is a free data retrieval call binding the contract method 0x0314d2b3.
//
// Solidity: function isGameFinalized(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameFinalized(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameFinalized", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameFinalized is a free data retrieval call binding the contract method 0x0314d2b3.
//
// Solidity: function isGameFinalized(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameFinalized(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameFinalized(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameFinalized is a free data retrieval call binding the contract method 0x0314d2b3.
//
// Solidity: function isGameFinalized(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameFinalized(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameFinalized(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameProper is a free data retrieval call binding the contract method 0x496b9c16.
//
// Solidity: function isGameProper(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameProper(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameProper", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameProper is a free data retrieval call binding the contract method 0x496b9c16.
//
// Solidity: function isGameProper(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameProper(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameProper(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameProper is a free data retrieval call binding the contract method 0x496b9c16.
//
// Solidity: function isGameProper(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameProper(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameProper(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRegistered is a free data retrieval call binding the contract method 0xee658e45.
//
// Solidity: function isGameRegistered(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameRegistered(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameRegistered", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameRegistered is a free data retrieval call binding the contract method 0xee658e45.
//
// Solidity: function isGameRegistered(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameRegistered(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRegistered(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRegistered is a free data retrieval call binding the contract method 0xee658e45.
//
// Solidity: function isGameRegistered(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameRegistered(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRegistered(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameResolved is a free data retrieval call binding the contract method 0xfdbb3dcf.
//
// Solidity: function isGameResolved(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameResolved(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameResolved", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameResolved is a free data retrieval call binding the contract method 0xfdbb3dcf.
//
// Solidity: function isGameResolved(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameResolved(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameResolved(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameResolved is a free data retrieval call binding the contract method 0xfdbb3dcf.
//
// Solidity: function isGameResolved(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameResolved(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameResolved(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRespected is a free data retrieval call binding the contract method 0x04e50fed.
//
// Solidity: function isGameRespected(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameRespected(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameRespected", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameRespected is a free data retrieval call binding the contract method 0x04e50fed.
//
// Solidity: function isGameRespected(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameRespected(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRespected(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRespected is a free data retrieval call binding the contract method 0x04e50fed.
//
// Solidity: function isGameRespected(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameRespected(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRespected(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRetired is a free data retrieval call binding the contract method 0x5958a193.
//
// Solidity: function isGameRetired(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) IsGameRetired(opts *bind.CallOpts, _game common.Address) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "isGameRetired", _game)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsGameRetired is a free data retrieval call binding the contract method 0x5958a193.
//
// Solidity: function isGameRetired(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) IsGameRetired(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRetired(&_AnchorStateRegistry.CallOpts, _game)
}

// IsGameRetired is a free data retrieval call binding the contract method 0x5958a193.
//
// Solidity: function isGameRetired(address _game) view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) IsGameRetired(_game common.Address) (bool, error) {
	return _AnchorStateRegistry.Contract.IsGameRetired(&_AnchorStateRegistry.CallOpts, _game)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) Paused(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "paused")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistrySession) Paused() (bool, error) {
	return _AnchorStateRegistry.Contract.Paused(&_AnchorStateRegistry.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(bool)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) Paused() (bool, error) {
	return _AnchorStateRegistry.Contract.Paused(&_AnchorStateRegistry.CallOpts)
}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) ProxyAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "proxyAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) ProxyAdmin() (common.Address, error) {
	return _AnchorStateRegistry.Contract.ProxyAdmin(&_AnchorStateRegistry.CallOpts)
}

// ProxyAdmin is a free data retrieval call binding the contract method 0x3e47158c.
//
// Solidity: function proxyAdmin() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) ProxyAdmin() (common.Address, error) {
	return _AnchorStateRegistry.Contract.ProxyAdmin(&_AnchorStateRegistry.CallOpts)
}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) ProxyAdminOwner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "proxyAdminOwner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) ProxyAdminOwner() (common.Address, error) {
	return _AnchorStateRegistry.Contract.ProxyAdminOwner(&_AnchorStateRegistry.CallOpts)
}

// ProxyAdminOwner is a free data retrieval call binding the contract method 0xdad544e0.
//
// Solidity: function proxyAdminOwner() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) ProxyAdminOwner() (common.Address, error) {
	return _AnchorStateRegistry.Contract.ProxyAdminOwner(&_AnchorStateRegistry.CallOpts)
}

// RespectedGameType is a free data retrieval call binding the contract method 0x3c9f397c.
//
// Solidity: function respectedGameType() view returns(uint32)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) RespectedGameType(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "respectedGameType")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// RespectedGameType is a free data retrieval call binding the contract method 0x3c9f397c.
//
// Solidity: function respectedGameType() view returns(uint32)
func (_AnchorStateRegistry *AnchorStateRegistrySession) RespectedGameType() (uint32, error) {
	return _AnchorStateRegistry.Contract.RespectedGameType(&_AnchorStateRegistry.CallOpts)
}

// RespectedGameType is a free data retrieval call binding the contract method 0x3c9f397c.
//
// Solidity: function respectedGameType() view returns(uint32)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) RespectedGameType() (uint32, error) {
	return _AnchorStateRegistry.Contract.RespectedGameType(&_AnchorStateRegistry.CallOpts)
}

// RetirementTimestamp is a free data retrieval call binding the contract method 0x4086d183.
//
// Solidity: function retirementTimestamp() view returns(uint64)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) RetirementTimestamp(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "retirementTimestamp")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// RetirementTimestamp is a free data retrieval call binding the contract method 0x4086d183.
//
// Solidity: function retirementTimestamp() view returns(uint64)
func (_AnchorStateRegistry *AnchorStateRegistrySession) RetirementTimestamp() (uint64, error) {
	return _AnchorStateRegistry.Contract.RetirementTimestamp(&_AnchorStateRegistry.CallOpts)
}

// RetirementTimestamp is a free data retrieval call binding the contract method 0x4086d183.
//
// Solidity: function retirementTimestamp() view returns(uint64)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) RetirementTimestamp() (uint64, error) {
	return _AnchorStateRegistry.Contract.RetirementTimestamp(&_AnchorStateRegistry.CallOpts)
}

// SuperchainConfig is a free data retrieval call binding the contract method 0x35e80ab3.
//
// Solidity: function superchainConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) SuperchainConfig(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "superchainConfig")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SuperchainConfig is a free data retrieval call binding the contract method 0x35e80ab3.
//
// Solidity: function superchainConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) SuperchainConfig() (common.Address, error) {
	return _AnchorStateRegistry.Contract.SuperchainConfig(&_AnchorStateRegistry.CallOpts)
}

// SuperchainConfig is a free data retrieval call binding the contract method 0x35e80ab3.
//
// Solidity: function superchainConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) SuperchainConfig() (common.Address, error) {
	return _AnchorStateRegistry.Contract.SuperchainConfig(&_AnchorStateRegistry.CallOpts)
}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) SystemConfig(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "systemConfig")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistrySession) SystemConfig() (common.Address, error) {
	return _AnchorStateRegistry.Contract.SystemConfig(&_AnchorStateRegistry.CallOpts)
}

// SystemConfig is a free data retrieval call binding the contract method 0x33d7e2bd.
//
// Solidity: function systemConfig() view returns(address)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) SystemConfig() (common.Address, error) {
	return _AnchorStateRegistry.Contract.SystemConfig(&_AnchorStateRegistry.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_AnchorStateRegistry *AnchorStateRegistryCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _AnchorStateRegistry.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_AnchorStateRegistry *AnchorStateRegistrySession) Version() (string, error) {
	return _AnchorStateRegistry.Contract.Version(&_AnchorStateRegistry.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_AnchorStateRegistry *AnchorStateRegistryCallerSession) Version() (string, error) {
	return _AnchorStateRegistry.Contract.Version(&_AnchorStateRegistry.CallOpts)
}

// BlacklistDisputeGame is a paid mutator transaction binding the contract method 0x7d6be8dc.
//
// Solidity: function blacklistDisputeGame(address _disputeGame) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactor) BlacklistDisputeGame(opts *bind.TransactOpts, _disputeGame common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.contract.Transact(opts, "blacklistDisputeGame", _disputeGame)
}

// BlacklistDisputeGame is a paid mutator transaction binding the contract method 0x7d6be8dc.
//
// Solidity: function blacklistDisputeGame(address _disputeGame) returns()
func (_AnchorStateRegistry *AnchorStateRegistrySession) BlacklistDisputeGame(_disputeGame common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.BlacklistDisputeGame(&_AnchorStateRegistry.TransactOpts, _disputeGame)
}

// BlacklistDisputeGame is a paid mutator transaction binding the contract method 0x7d6be8dc.
//
// Solidity: function blacklistDisputeGame(address _disputeGame) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactorSession) BlacklistDisputeGame(_disputeGame common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.BlacklistDisputeGame(&_AnchorStateRegistry.TransactOpts, _disputeGame)
}

// Initialize is a paid mutator transaction binding the contract method 0x47a222c5.
//
// Solidity: function initialize(address _systemConfig, address _disputeGameFactory, (bytes32,uint256) _startingAnchorRoot, uint32 _startingRespectedGameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactor) Initialize(opts *bind.TransactOpts, _systemConfig common.Address, _disputeGameFactory common.Address, _startingAnchorRoot Proposal, _startingRespectedGameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.contract.Transact(opts, "initialize", _systemConfig, _disputeGameFactory, _startingAnchorRoot, _startingRespectedGameType)
}

// Initialize is a paid mutator transaction binding the contract method 0x47a222c5.
//
// Solidity: function initialize(address _systemConfig, address _disputeGameFactory, (bytes32,uint256) _startingAnchorRoot, uint32 _startingRespectedGameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistrySession) Initialize(_systemConfig common.Address, _disputeGameFactory common.Address, _startingAnchorRoot Proposal, _startingRespectedGameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.Initialize(&_AnchorStateRegistry.TransactOpts, _systemConfig, _disputeGameFactory, _startingAnchorRoot, _startingRespectedGameType)
}

// Initialize is a paid mutator transaction binding the contract method 0x47a222c5.
//
// Solidity: function initialize(address _systemConfig, address _disputeGameFactory, (bytes32,uint256) _startingAnchorRoot, uint32 _startingRespectedGameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactorSession) Initialize(_systemConfig common.Address, _disputeGameFactory common.Address, _startingAnchorRoot Proposal, _startingRespectedGameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.Initialize(&_AnchorStateRegistry.TransactOpts, _systemConfig, _disputeGameFactory, _startingAnchorRoot, _startingRespectedGameType)
}

// SetAnchorState is a paid mutator transaction binding the contract method 0x17cf21a9.
//
// Solidity: function setAnchorState(address _game) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactor) SetAnchorState(opts *bind.TransactOpts, _game common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.contract.Transact(opts, "setAnchorState", _game)
}

// SetAnchorState is a paid mutator transaction binding the contract method 0x17cf21a9.
//
// Solidity: function setAnchorState(address _game) returns()
func (_AnchorStateRegistry *AnchorStateRegistrySession) SetAnchorState(_game common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.SetAnchorState(&_AnchorStateRegistry.TransactOpts, _game)
}

// SetAnchorState is a paid mutator transaction binding the contract method 0x17cf21a9.
//
// Solidity: function setAnchorState(address _game) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactorSession) SetAnchorState(_game common.Address) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.SetAnchorState(&_AnchorStateRegistry.TransactOpts, _game)
}

// SetRespectedGameType is a paid mutator transaction binding the contract method 0x7fc48504.
//
// Solidity: function setRespectedGameType(uint32 _gameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactor) SetRespectedGameType(opts *bind.TransactOpts, _gameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.contract.Transact(opts, "setRespectedGameType", _gameType)
}

// SetRespectedGameType is a paid mutator transaction binding the contract method 0x7fc48504.
//
// Solidity: function setRespectedGameType(uint32 _gameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistrySession) SetRespectedGameType(_gameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.SetRespectedGameType(&_AnchorStateRegistry.TransactOpts, _gameType)
}

// SetRespectedGameType is a paid mutator transaction binding the contract method 0x7fc48504.
//
// Solidity: function setRespectedGameType(uint32 _gameType) returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactorSession) SetRespectedGameType(_gameType uint32) (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.SetRespectedGameType(&_AnchorStateRegistry.TransactOpts, _gameType)
}

// UpdateRetirementTimestamp is a paid mutator transaction binding the contract method 0xd5a3e12e.
//
// Solidity: function updateRetirementTimestamp() returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactor) UpdateRetirementTimestamp(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AnchorStateRegistry.contract.Transact(opts, "updateRetirementTimestamp")
}

// UpdateRetirementTimestamp is a paid mutator transaction binding the contract method 0xd5a3e12e.
//
// Solidity: function updateRetirementTimestamp() returns()
func (_AnchorStateRegistry *AnchorStateRegistrySession) UpdateRetirementTimestamp() (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.UpdateRetirementTimestamp(&_AnchorStateRegistry.TransactOpts)
}

// UpdateRetirementTimestamp is a paid mutator transaction binding the contract method 0xd5a3e12e.
//
// Solidity: function updateRetirementTimestamp() returns()
func (_AnchorStateRegistry *AnchorStateRegistryTransactorSession) UpdateRetirementTimestamp() (*types.Transaction, error) {
	return _AnchorStateRegistry.Contract.UpdateRetirementTimestamp(&_AnchorStateRegistry.TransactOpts)
}

// AnchorStateRegistryAnchorUpdatedIterator is returned from FilterAnchorUpdated and is used to iterate over the raw logs and unpacked data for AnchorUpdated events raised by the AnchorStateRegistry contract.
type AnchorStateRegistryAnchorUpdatedIterator struct {
	Event *AnchorStateRegistryAnchorUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnchorStateRegistryAnchorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorStateRegistryAnchorUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnchorStateRegistryAnchorUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnchorStateRegistryAnchorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorStateRegistryAnchorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorStateRegistryAnchorUpdated represents a AnchorUpdated event raised by the AnchorStateRegistry contract.
type AnchorStateRegistryAnchorUpdated struct {
	Game common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterAnchorUpdated is a free log retrieval operation binding the contract event 0x474f180d74ea8751955ee261c93ff8270411b180408d1014c49f552c92a4d11e.
//
// Solidity: event AnchorUpdated(address indexed game)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) FilterAnchorUpdated(opts *bind.FilterOpts, game []common.Address) (*AnchorStateRegistryAnchorUpdatedIterator, error) {

	var gameRule []interface{}
	for _, gameItem := range game {
		gameRule = append(gameRule, gameItem)
	}

	logs, sub, err := _AnchorStateRegistry.contract.FilterLogs(opts, "AnchorUpdated", gameRule)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryAnchorUpdatedIterator{contract: _AnchorStateRegistry.contract, event: "AnchorUpdated", logs: logs, sub: sub}, nil
}

// WatchAnchorUpdated is a free log subscription operation binding the contract event 0x474f180d74ea8751955ee261c93ff8270411b180408d1014c49f552c92a4d11e.
//
// Solidity: event AnchorUpdated(address indexed game)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) WatchAnchorUpdated(opts *bind.WatchOpts, sink chan<- *AnchorStateRegistryAnchorUpdated, game []common.Address) (event.Subscription, error) {

	var gameRule []interface{}
	for _, gameItem := range game {
		gameRule = append(gameRule, gameItem)
	}

	logs, sub, err := _AnchorStateRegistry.contract.WatchLogs(opts, "AnchorUpdated", gameRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorStateRegistryAnchorUpdated)
				if err := _AnchorStateRegistry.contract.UnpackLog(event, "AnchorUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAnchorUpdated is a log parse operation binding the contract event 0x474f180d74ea8751955ee261c93ff8270411b180408d1014c49f552c92a4d11e.
//
// Solidity: event AnchorUpdated(address indexed game)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) ParseAnchorUpdated(log types.Log) (*AnchorStateRegistryAnchorUpdated, error) {
	event := new(AnchorStateRegistryAnchorUpdated)
	if err := _AnchorStateRegistry.contract.UnpackLog(event, "AnchorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnchorStateRegistryDisputeGameBlacklistedIterator is returned from FilterDisputeGameBlacklisted and is used to iterate over the raw logs and unpacked data for DisputeGameBlacklisted events raised by the AnchorStateRegistry contract.
type AnchorStateRegistryDisputeGameBlacklistedIterator struct {
	Event *AnchorStateRegistryDisputeGameBlacklisted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnchorStateRegistryDisputeGameBlacklistedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorStateRegistryDisputeGameBlacklisted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnchorStateRegistryDisputeGameBlacklisted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnchorStateRegistryDisputeGameBlacklistedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorStateRegistryDisputeGameBlacklistedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorStateRegistryDisputeGameBlacklisted represents a DisputeGameBlacklisted event raised by the AnchorStateRegistry contract.
type AnchorStateRegistryDisputeGameBlacklisted struct {
	DisputeGame common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterDisputeGameBlacklisted is a free log retrieval operation binding the contract event 0x192c289026d59a41a27f5aea08f3969b57931b0589202d14f4368cded95d3cda.
//
// Solidity: event DisputeGameBlacklisted(address indexed disputeGame)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) FilterDisputeGameBlacklisted(opts *bind.FilterOpts, disputeGame []common.Address) (*AnchorStateRegistryDisputeGameBlacklistedIterator, error) {

	var disputeGameRule []interface{}
	for _, disputeGameItem := range disputeGame {
		disputeGameRule = append(disputeGameRule, disputeGameItem)
	}

	logs, sub, err := _AnchorStateRegistry.contract.FilterLogs(opts, "DisputeGameBlacklisted", disputeGameRule)
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryDisputeGameBlacklistedIterator{contract: _AnchorStateRegistry.contract, event: "DisputeGameBlacklisted", logs: logs, sub: sub}, nil
}

// WatchDisputeGameBlacklisted is a free log subscription operation binding the contract event 0x192c289026d59a41a27f5aea08f3969b57931b0589202d14f4368cded95d3cda.
//
// Solidity: event DisputeGameBlacklisted(address indexed disputeGame)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) WatchDisputeGameBlacklisted(opts *bind.WatchOpts, sink chan<- *AnchorStateRegistryDisputeGameBlacklisted, disputeGame []common.Address) (event.Subscription, error) {

	var disputeGameRule []interface{}
	for _, disputeGameItem := range disputeGame {
		disputeGameRule = append(disputeGameRule, disputeGameItem)
	}

	logs, sub, err := _AnchorStateRegistry.contract.WatchLogs(opts, "DisputeGameBlacklisted", disputeGameRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorStateRegistryDisputeGameBlacklisted)
				if err := _AnchorStateRegistry.contract.UnpackLog(event, "DisputeGameBlacklisted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDisputeGameBlacklisted is a log parse operation binding the contract event 0x192c289026d59a41a27f5aea08f3969b57931b0589202d14f4368cded95d3cda.
//
// Solidity: event DisputeGameBlacklisted(address indexed disputeGame)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) ParseDisputeGameBlacklisted(log types.Log) (*AnchorStateRegistryDisputeGameBlacklisted, error) {
	event := new(AnchorStateRegistryDisputeGameBlacklisted)
	if err := _AnchorStateRegistry.contract.UnpackLog(event, "DisputeGameBlacklisted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnchorStateRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AnchorStateRegistry contract.
type AnchorStateRegistryInitializedIterator struct {
	Event *AnchorStateRegistryInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnchorStateRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorStateRegistryInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnchorStateRegistryInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnchorStateRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorStateRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorStateRegistryInitialized represents a Initialized event raised by the AnchorStateRegistry contract.
type AnchorStateRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*AnchorStateRegistryInitializedIterator, error) {

	logs, sub, err := _AnchorStateRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryInitializedIterator{contract: _AnchorStateRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AnchorStateRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _AnchorStateRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorStateRegistryInitialized)
				if err := _AnchorStateRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) ParseInitialized(log types.Log) (*AnchorStateRegistryInitialized, error) {
	event := new(AnchorStateRegistryInitialized)
	if err := _AnchorStateRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnchorStateRegistryRespectedGameTypeSetIterator is returned from FilterRespectedGameTypeSet and is used to iterate over the raw logs and unpacked data for RespectedGameTypeSet events raised by the AnchorStateRegistry contract.
type AnchorStateRegistryRespectedGameTypeSetIterator struct {
	Event *AnchorStateRegistryRespectedGameTypeSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnchorStateRegistryRespectedGameTypeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorStateRegistryRespectedGameTypeSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnchorStateRegistryRespectedGameTypeSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnchorStateRegistryRespectedGameTypeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorStateRegistryRespectedGameTypeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorStateRegistryRespectedGameTypeSet represents a RespectedGameTypeSet event raised by the AnchorStateRegistry contract.
type AnchorStateRegistryRespectedGameTypeSet struct {
	GameType uint32
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterRespectedGameTypeSet is a free log retrieval operation binding the contract event 0xcee0703b5e4bad4efededab85c9fd1aec17dee7c5f6c584330e0509b677745a2.
//
// Solidity: event RespectedGameTypeSet(uint32 gameType)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) FilterRespectedGameTypeSet(opts *bind.FilterOpts) (*AnchorStateRegistryRespectedGameTypeSetIterator, error) {

	logs, sub, err := _AnchorStateRegistry.contract.FilterLogs(opts, "RespectedGameTypeSet")
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryRespectedGameTypeSetIterator{contract: _AnchorStateRegistry.contract, event: "RespectedGameTypeSet", logs: logs, sub: sub}, nil
}

// WatchRespectedGameTypeSet is a free log subscription operation binding the contract event 0xcee0703b5e4bad4efededab85c9fd1aec17dee7c5f6c584330e0509b677745a2.
//
// Solidity: event RespectedGameTypeSet(uint32 gameType)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) WatchRespectedGameTypeSet(opts *bind.WatchOpts, sink chan<- *AnchorStateRegistryRespectedGameTypeSet) (event.Subscription, error) {

	logs, sub, err := _AnchorStateRegistry.contract.WatchLogs(opts, "RespectedGameTypeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorStateRegistryRespectedGameTypeSet)
				if err := _AnchorStateRegistry.contract.UnpackLog(event, "RespectedGameTypeSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRespectedGameTypeSet is a log parse operation binding the contract event 0xcee0703b5e4bad4efededab85c9fd1aec17dee7c5f6c584330e0509b677745a2.
//
// Solidity: event RespectedGameTypeSet(uint32 gameType)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) ParseRespectedGameTypeSet(log types.Log) (*AnchorStateRegistryRespectedGameTypeSet, error) {
	event := new(AnchorStateRegistryRespectedGameTypeSet)
	if err := _AnchorStateRegistry.contract.UnpackLog(event, "RespectedGameTypeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AnchorStateRegistryRetirementTimestampSetIterator is returned from FilterRetirementTimestampSet and is used to iterate over the raw logs and unpacked data for RetirementTimestampSet events raised by the AnchorStateRegistry contract.
type AnchorStateRegistryRetirementTimestampSetIterator struct {
	Event *AnchorStateRegistryRetirementTimestampSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *AnchorStateRegistryRetirementTimestampSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AnchorStateRegistryRetirementTimestampSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(AnchorStateRegistryRetirementTimestampSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *AnchorStateRegistryRetirementTimestampSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AnchorStateRegistryRetirementTimestampSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AnchorStateRegistryRetirementTimestampSet represents a RetirementTimestampSet event raised by the AnchorStateRegistry contract.
type AnchorStateRegistryRetirementTimestampSet struct {
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterRetirementTimestampSet is a free log retrieval operation binding the contract event 0x6e5b1ba771e8e484f741ed085f039ff4e5c6e882eaf68f550fb390922d0ae4a7.
//
// Solidity: event RetirementTimestampSet(uint256 timestamp)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) FilterRetirementTimestampSet(opts *bind.FilterOpts) (*AnchorStateRegistryRetirementTimestampSetIterator, error) {

	logs, sub, err := _AnchorStateRegistry.contract.FilterLogs(opts, "RetirementTimestampSet")
	if err != nil {
		return nil, err
	}
	return &AnchorStateRegistryRetirementTimestampSetIterator{contract: _AnchorStateRegistry.contract, event: "RetirementTimestampSet", logs: logs, sub: sub}, nil
}

// WatchRetirementTimestampSet is a free log subscription operation binding the contract event 0x6e5b1ba771e8e484f741ed085f039ff4e5c6e882eaf68f550fb390922d0ae4a7.
//
// Solidity: event RetirementTimestampSet(uint256 timestamp)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) WatchRetirementTimestampSet(opts *bind.WatchOpts, sink chan<- *AnchorStateRegistryRetirementTimestampSet) (event.Subscription, error) {

	logs, sub, err := _AnchorStateRegistry.contract.WatchLogs(opts, "RetirementTimestampSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AnchorStateRegistryRetirementTimestampSet)
				if err := _AnchorStateRegistry.contract.UnpackLog(event, "RetirementTimestampSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRetirementTimestampSet is a log parse operation binding the contract event 0x6e5b1ba771e8e484f741ed085f039ff4e5c6e882eaf68f550fb390922d0ae4a7.
//
// Solidity: event RetirementTimestampSet(uint256 timestamp)
func (_AnchorStateRegistry *AnchorStateRegistryFilterer) ParseRetirementTimestampSet(log types.Log) (*AnchorStateRegistryRetirementTimestampSet, error) {
	event := new(AnchorStateRegistryRetirementTimestampSet)
	if err := _AnchorStateRegistry.contract.UnpackLog(event, "RetirementTimestampSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
