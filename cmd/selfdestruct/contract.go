// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package main

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

// SelfdestructerMetaData contains all meta data concerning the Selfdestructer contract.
var SelfdestructerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"Destruct\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Size\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b50610269806100206000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c806342e90c33146100465780637e1a6753146100505780637f225bf71461005a575b600080fd5b61004e610078565b005b6100586100c9565b005b6100626100e2565b60405161006f91906100fb565b60405180910390f35b60004490505b61ea605a11156100965780815560018101905061007e565b602044826100a491906101c6565b6100ae919061016c565b600160008282546100bf9190610116565b9250508190555050565b3373ffffffffffffffffffffffffffffffffffffffff16ff5b6000600154905090565b6100f5816101fa565b82525050565b600060208201905061011060008301846100ec565b92915050565b6000610121826101fa565b915061012c836101fa565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561016157610160610204565b5b828201905092915050565b6000610177826101fa565b9150610182836101fa565b9250817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04831182151516156101bb576101ba610204565b5b828202905092915050565b60006101d1826101fa565b91506101dc836101fa565b9250828210156101ef576101ee610204565b5b828203905092915050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fdfea26469706673582212201fe8a625b5007fbacbe4612f63b962cbb26d10c1a91963edca69e925796eee3f64736f6c63430008070033",
}

// SelfdestructerABI is the input ABI used to generate the binding from.
// Deprecated: Use SelfdestructerMetaData.ABI instead.
var SelfdestructerABI = SelfdestructerMetaData.ABI

// SelfdestructerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SelfdestructerMetaData.Bin instead.
var SelfdestructerBin = SelfdestructerMetaData.Bin

// DeploySelfdestructer deploys a new Ethereum contract, binding an instance of Selfdestructer to it.
func DeploySelfdestructer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Selfdestructer, error) {
	parsed, err := SelfdestructerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SelfdestructerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Selfdestructer{SelfdestructerCaller: SelfdestructerCaller{contract: contract}, SelfdestructerTransactor: SelfdestructerTransactor{contract: contract}, SelfdestructerFilterer: SelfdestructerFilterer{contract: contract}}, nil
}

// Selfdestructer is an auto generated Go binding around an Ethereum contract.
type Selfdestructer struct {
	SelfdestructerCaller     // Read-only binding to the contract
	SelfdestructerTransactor // Write-only binding to the contract
	SelfdestructerFilterer   // Log filterer for contract events
}

// SelfdestructerCaller is an auto generated read-only Go binding around an Ethereum contract.
type SelfdestructerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfdestructerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SelfdestructerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfdestructerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SelfdestructerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SelfdestructerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SelfdestructerSession struct {
	Contract     *Selfdestructer   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SelfdestructerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SelfdestructerCallerSession struct {
	Contract *SelfdestructerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// SelfdestructerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SelfdestructerTransactorSession struct {
	Contract     *SelfdestructerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// SelfdestructerRaw is an auto generated low-level Go binding around an Ethereum contract.
type SelfdestructerRaw struct {
	Contract *Selfdestructer // Generic contract binding to access the raw methods on
}

// SelfdestructerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SelfdestructerCallerRaw struct {
	Contract *SelfdestructerCaller // Generic read-only contract binding to access the raw methods on
}

// SelfdestructerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SelfdestructerTransactorRaw struct {
	Contract *SelfdestructerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSelfdestructer creates a new instance of Selfdestructer, bound to a specific deployed contract.
func NewSelfdestructer(address common.Address, backend bind.ContractBackend) (*Selfdestructer, error) {
	contract, err := bindSelfdestructer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Selfdestructer{SelfdestructerCaller: SelfdestructerCaller{contract: contract}, SelfdestructerTransactor: SelfdestructerTransactor{contract: contract}, SelfdestructerFilterer: SelfdestructerFilterer{contract: contract}}, nil
}

// NewSelfdestructerCaller creates a new read-only instance of Selfdestructer, bound to a specific deployed contract.
func NewSelfdestructerCaller(address common.Address, caller bind.ContractCaller) (*SelfdestructerCaller, error) {
	contract, err := bindSelfdestructer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SelfdestructerCaller{contract: contract}, nil
}

// NewSelfdestructerTransactor creates a new write-only instance of Selfdestructer, bound to a specific deployed contract.
func NewSelfdestructerTransactor(address common.Address, transactor bind.ContractTransactor) (*SelfdestructerTransactor, error) {
	contract, err := bindSelfdestructer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SelfdestructerTransactor{contract: contract}, nil
}

// NewSelfdestructerFilterer creates a new log filterer instance of Selfdestructer, bound to a specific deployed contract.
func NewSelfdestructerFilterer(address common.Address, filterer bind.ContractFilterer) (*SelfdestructerFilterer, error) {
	contract, err := bindSelfdestructer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SelfdestructerFilterer{contract: contract}, nil
}

// bindSelfdestructer binds a generic wrapper to an already deployed contract.
func bindSelfdestructer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SelfdestructerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Selfdestructer *SelfdestructerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Selfdestructer.Contract.SelfdestructerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Selfdestructer *SelfdestructerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Selfdestructer.Contract.SelfdestructerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Selfdestructer *SelfdestructerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Selfdestructer.Contract.SelfdestructerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Selfdestructer *SelfdestructerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Selfdestructer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Selfdestructer *SelfdestructerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Selfdestructer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Selfdestructer *SelfdestructerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Selfdestructer.Contract.contract.Transact(opts, method, params...)
}

// Size is a free data retrieval call binding the contract method 0x7f225bf7.
//
// Solidity: function Size() view returns(uint256)
func (_Selfdestructer *SelfdestructerCaller) Size(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Selfdestructer.contract.Call(opts, &out, "Size")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Size is a free data retrieval call binding the contract method 0x7f225bf7.
//
// Solidity: function Size() view returns(uint256)
func (_Selfdestructer *SelfdestructerSession) Size() (*big.Int, error) {
	return _Selfdestructer.Contract.Size(&_Selfdestructer.CallOpts)
}

// Size is a free data retrieval call binding the contract method 0x7f225bf7.
//
// Solidity: function Size() view returns(uint256)
func (_Selfdestructer *SelfdestructerCallerSession) Size() (*big.Int, error) {
	return _Selfdestructer.Contract.Size(&_Selfdestructer.CallOpts)
}

// Destruct is a paid mutator transaction binding the contract method 0x7e1a6753.
//
// Solidity: function Destruct() returns()
func (_Selfdestructer *SelfdestructerTransactor) Destruct(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Selfdestructer.contract.Transact(opts, "Destruct")
}

// Destruct is a paid mutator transaction binding the contract method 0x7e1a6753.
//
// Solidity: function Destruct() returns()
func (_Selfdestructer *SelfdestructerSession) Destruct() (*types.Transaction, error) {
	return _Selfdestructer.Contract.Destruct(&_Selfdestructer.TransactOpts)
}

// Destruct is a paid mutator transaction binding the contract method 0x7e1a6753.
//
// Solidity: function Destruct() returns()
func (_Selfdestructer *SelfdestructerTransactorSession) Destruct() (*types.Transaction, error) {
	return _Selfdestructer.Contract.Destruct(&_Selfdestructer.TransactOpts)
}

// Store is a paid mutator transaction binding the contract method 0x42e90c33.
//
// Solidity: function Store() returns()
func (_Selfdestructer *SelfdestructerTransactor) Store(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Selfdestructer.contract.Transact(opts, "Store")
}

// Store is a paid mutator transaction binding the contract method 0x42e90c33.
//
// Solidity: function Store() returns()
func (_Selfdestructer *SelfdestructerSession) Store() (*types.Transaction, error) {
	return _Selfdestructer.Contract.Store(&_Selfdestructer.TransactOpts)
}

// Store is a paid mutator transaction binding the contract method 0x42e90c33.
//
// Solidity: function Store() returns()
func (_Selfdestructer *SelfdestructerTransactorSession) Store() (*types.Transaction, error) {
	return _Selfdestructer.Contract.Store(&_Selfdestructer.TransactOpts)
}
