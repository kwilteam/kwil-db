// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abi

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

// EscrowMetaData contains all meta data concerning the Escrow contract.
var EscrowMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_escrowToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"receiver\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"nonce\",\"type\":\"string\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"}],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"validator\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amt\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"escrowToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"pools\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"nonce\",\"type\":\"string\"}],\"name\":\"returnDeposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b506040516200109d3803806200109d8339818101604052810190620000379190620000e8565b806000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550506200011a565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000620000b08262000083565b9050919050565b620000c281620000a3565b8114620000ce57600080fd5b50565b600081519050620000e281620000b7565b92915050565b6000602082840312156200010157620001006200007e565b5b60006200011184828501620000d1565b91505092915050565b610f73806200012a6000396000f3fe60806040526004361061004a5760003560e01c806305da28201461004f5780632fe319da1461007857806347e7ef24146100a3578063901754d7146100bf578063b203bb99146100fc575b600080fd5b34801561005b57600080fd5b506100766004803603810190610071919061093e565b610139565b005b34801561008457600080fd5b5061008d6104ba565b60405161009a9190610a20565b60405180910390f35b6100bd60048036038101906100b89190610a3b565b6104de565b005b3480156100cb57600080fd5b506100e660048036038101906100e19190610a7b565b6106a4565b6040516100f39190610aca565b60405180910390f35b34801561010857600080fd5b50610123600480360381019061011e9190610a7b565b6106c9565b6040516101309190610aca565b60405180910390f35b600082846101479190610b14565b905080600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020541015610208576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101ff90610ba5565b60405180910390fd5b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb86866040518363ffffffff1660e01b8152600401610263929190610bd4565b6020604051808303816000875af1158015610282573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906102a69190610c35565b6102e5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102dc90610cd4565b60405180910390fd5b60008311156103cc5760008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb33856040518363ffffffff1660e01b8152600401610349929190610bd4565b6020604051808303816000875af1158015610368573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061038c9190610c35565b6103cb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103c290610d66565b60405180910390fd5b5b80600160003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008773ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546104589190610d86565b925050819055503373ffffffffffffffffffffffffffffffffffffffff167ffbe316db8c3cfdc314d06e8ab0b8baf61d03bd5aa62893c764ef1215a962d268868686866040516104ab9493929190610e28565b60405180910390a25050505050565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60008054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3330846040518463ffffffff1660e01b815260040161053b93929190610e74565b6020604051808303816000875af115801561055a573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061057e9190610c35565b6105bd576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105b490610f1d565b60405180910390fd5b80600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546106499190610b14565b925050819055508173ffffffffffffffffffffffffffffffffffffffff167f5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f623383604051610698929190610bd4565b60405180910390a25050565b6001602052816000526040600020602052806000526040600020600091509150505481565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905092915050565b6000604051905090565b600080fd5b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061078f82610764565b9050919050565b61079f81610784565b81146107aa57600080fd5b50565b6000813590506107bc81610796565b92915050565b6000819050919050565b6107d5816107c2565b81146107e057600080fd5b50565b6000813590506107f2816107cc565b92915050565b600080fd5b600080fd5b6000601f19601f8301169050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b61084b82610802565b810181811067ffffffffffffffff8211171561086a57610869610813565b5b80604052505050565b600061087d610750565b90506108898282610842565b919050565b600067ffffffffffffffff8211156108a9576108a8610813565b5b6108b282610802565b9050602081019050919050565b82818337600083830152505050565b60006108e16108dc8461088e565b610873565b9050828152602081018484840111156108fd576108fc6107fd565b5b6109088482856108bf565b509392505050565b600082601f830112610925576109246107f8565b5b81356109358482602086016108ce565b91505092915050565b600080600080608085870312156109585761095761075a565b5b6000610966878288016107ad565b9450506020610977878288016107e3565b9350506040610988878288016107e3565b925050606085013567ffffffffffffffff8111156109a9576109a861075f565b5b6109b587828801610910565b91505092959194509250565b6000819050919050565b60006109e66109e16109dc84610764565b6109c1565b610764565b9050919050565b60006109f8826109cb565b9050919050565b6000610a0a826109ed565b9050919050565b610a1a816109ff565b82525050565b6000602082019050610a356000830184610a11565b92915050565b60008060408385031215610a5257610a5161075a565b5b6000610a60858286016107ad565b9250506020610a71858286016107e3565b9150509250929050565b60008060408385031215610a9257610a9161075a565b5b6000610aa0858286016107ad565b9250506020610ab1858286016107ad565b9150509250929050565b610ac4816107c2565b82525050565b6000602082019050610adf6000830184610abb565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610b1f826107c2565b9150610b2a836107c2565b9250828201905080821115610b4257610b41610ae5565b5b92915050565b600082825260208201905092915050565b7f4e6f7420656e6f75676820746f207472616e73666572206261636b0000000000600082015250565b6000610b8f601b83610b48565b9150610b9a82610b59565b602082019050919050565b60006020820190508181036000830152610bbe81610b82565b9050919050565b610bce81610784565b82525050565b6000604082019050610be96000830185610bc5565b610bf66020830184610abb565b9392505050565b60008115159050919050565b610c1281610bfd565b8114610c1d57600080fd5b50565b600081519050610c2f81610c09565b92915050565b600060208284031215610c4b57610c4a61075a565b5b6000610c5984828501610c20565b91505092915050565b7f436f756c64206e6f74207472616e736665722066756e6473206261636b20746f60008201527f206f776e65720000000000000000000000000000000000000000000000000000602082015250565b6000610cbe602683610b48565b9150610cc982610c62565b604082019050919050565b60006020820190508181036000830152610ced81610cb1565b9050919050565b7f436f756c64206e6f74207472616e736665722066756e647320746f2076616c6960008201527f6461746f72000000000000000000000000000000000000000000000000000000602082015250565b6000610d50602583610b48565b9150610d5b82610cf4565b604082019050919050565b60006020820190508181036000830152610d7f81610d43565b9050919050565b6000610d91826107c2565b9150610d9c836107c2565b9250828203905081811115610db457610db3610ae5565b5b92915050565b600081519050919050565b60005b83811015610de3578082015181840152602081019050610dc8565b60008484015250505050565b6000610dfa82610dba565b610e048185610b48565b9350610e14818560208601610dc5565b610e1d81610802565b840191505092915050565b6000608082019050610e3d6000830187610bc5565b610e4a6020830186610abb565b610e576040830185610abb565b8181036060830152610e698184610def565b905095945050505050565b6000606082019050610e896000830186610bc5565b610e966020830185610bc5565b610ea36040830184610abb565b949350505050565b7f4465706f736974206661696c65643a20746f6b656e20646964206e6f7420737560008201527f636365737366756c6c79207472616e7366657200000000000000000000000000602082015250565b6000610f07603383610b48565b9150610f1282610eab565b604082019050919050565b60006020820190508181036000830152610f3681610efa565b905091905056fea2646970667358221220f91d9524e2cd11bcfb8a19e1e917bf267311dcbf050115c98516120365e55f1d64736f6c63430008110033",
}

// EscrowABI is the input ABI used to generate the binding from.
// Deprecated: Use EscrowMetaData.ABI instead.
var EscrowABI = EscrowMetaData.ABI

// EscrowBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use EscrowMetaData.Bin instead.
var EscrowBin = EscrowMetaData.Bin

// DeployEscrow deploys a new Ethereum contract, binding an instance of Escrow to it.
func DeployEscrow(auth *bind.TransactOpts, backend bind.ContractBackend, _escrowToken common.Address) (common.Address, *types.Transaction, *Escrow, error) {
	parsed, err := EscrowMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(EscrowBin), backend, _escrowToken)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Escrow{EscrowCaller: EscrowCaller{contract: contract}, EscrowTransactor: EscrowTransactor{contract: contract}, EscrowFilterer: EscrowFilterer{contract: contract}}, nil
}

// Escrow is an auto generated Go binding around an Ethereum contract.
type Escrow struct {
	EscrowCaller     // Read-only binding to the contract
	EscrowTransactor // Write-only binding to the contract
	EscrowFilterer   // Log filterer for contract events
}

// EscrowCaller is an auto generated read-only Go binding around an Ethereum contract.
type EscrowCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EscrowTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EscrowFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EscrowSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EscrowSession struct {
	Contract     *Escrow           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EscrowCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EscrowCallerSession struct {
	Contract *EscrowCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// EscrowTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EscrowTransactorSession struct {
	Contract     *EscrowTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EscrowRaw is an auto generated low-level Go binding around an Ethereum contract.
type EscrowRaw struct {
	Contract *Escrow // Generic contract binding to access the raw methods on
}

// EscrowCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EscrowCallerRaw struct {
	Contract *EscrowCaller // Generic read-only contract binding to access the raw methods on
}

// EscrowTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EscrowTransactorRaw struct {
	Contract *EscrowTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEscrow creates a new instance of Escrow, bound to a specific deployed contract.
func NewEscrow(address common.Address, backend bind.ContractBackend) (*Escrow, error) {
	contract, err := bindEscrow(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Escrow{EscrowCaller: EscrowCaller{contract: contract}, EscrowTransactor: EscrowTransactor{contract: contract}, EscrowFilterer: EscrowFilterer{contract: contract}}, nil
}

// NewEscrowCaller creates a new read-only instance of Escrow, bound to a specific deployed contract.
func NewEscrowCaller(address common.Address, caller bind.ContractCaller) (*EscrowCaller, error) {
	contract, err := bindEscrow(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EscrowCaller{contract: contract}, nil
}

// NewEscrowTransactor creates a new write-only instance of Escrow, bound to a specific deployed contract.
func NewEscrowTransactor(address common.Address, transactor bind.ContractTransactor) (*EscrowTransactor, error) {
	contract, err := bindEscrow(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EscrowTransactor{contract: contract}, nil
}

// NewEscrowFilterer creates a new log filterer instance of Escrow, bound to a specific deployed contract.
func NewEscrowFilterer(address common.Address, filterer bind.ContractFilterer) (*EscrowFilterer, error) {
	contract, err := bindEscrow(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EscrowFilterer{contract: contract}, nil
}

// bindEscrow binds a generic wrapper to an already deployed contract.
func bindEscrow(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EscrowMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Escrow *EscrowRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Escrow.Contract.EscrowCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Escrow *EscrowRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Escrow.Contract.EscrowTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Escrow *EscrowRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Escrow.Contract.EscrowTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Escrow *EscrowCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Escrow.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Escrow *EscrowTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Escrow.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Escrow *EscrowTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Escrow.Contract.contract.Transact(opts, method, params...)
}

// Balance is a free data retrieval call binding the contract method 0xb203bb99.
//
// Solidity: function balance(address wallet, address validator) view returns(uint256)
func (_Escrow *EscrowCaller) Balance(opts *bind.CallOpts, wallet common.Address, validator common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "balance", wallet, validator)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xb203bb99.
//
// Solidity: function balance(address wallet, address validator) view returns(uint256)
func (_Escrow *EscrowSession) Balance(wallet common.Address, validator common.Address) (*big.Int, error) {
	return _Escrow.Contract.Balance(&_Escrow.CallOpts, wallet, validator)
}

// Balance is a free data retrieval call binding the contract method 0xb203bb99.
//
// Solidity: function balance(address wallet, address validator) view returns(uint256)
func (_Escrow *EscrowCallerSession) Balance(wallet common.Address, validator common.Address) (*big.Int, error) {
	return _Escrow.Contract.Balance(&_Escrow.CallOpts, wallet, validator)
}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Escrow *EscrowCaller) EscrowToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "escrowToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Escrow *EscrowSession) EscrowToken() (common.Address, error) {
	return _Escrow.Contract.EscrowToken(&_Escrow.CallOpts)
}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Escrow *EscrowCallerSession) EscrowToken() (common.Address, error) {
	return _Escrow.Contract.EscrowToken(&_Escrow.CallOpts)
}

// Pools is a free data retrieval call binding the contract method 0x901754d7.
//
// Solidity: function pools(address , address ) view returns(uint256)
func (_Escrow *EscrowCaller) Pools(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "pools", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Pools is a free data retrieval call binding the contract method 0x901754d7.
//
// Solidity: function pools(address , address ) view returns(uint256)
func (_Escrow *EscrowSession) Pools(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Escrow.Contract.Pools(&_Escrow.CallOpts, arg0, arg1)
}

// Pools is a free data retrieval call binding the contract method 0x901754d7.
//
// Solidity: function pools(address , address ) view returns(uint256)
func (_Escrow *EscrowCallerSession) Pools(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Escrow.Contract.Pools(&_Escrow.CallOpts, arg0, arg1)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address validator, uint256 amt) payable returns()
func (_Escrow *EscrowTransactor) Deposit(opts *bind.TransactOpts, validator common.Address, amt *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "deposit", validator, amt)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address validator, uint256 amt) payable returns()
func (_Escrow *EscrowSession) Deposit(validator common.Address, amt *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.Deposit(&_Escrow.TransactOpts, validator, amt)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address validator, uint256 amt) payable returns()
func (_Escrow *EscrowTransactorSession) Deposit(validator common.Address, amt *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.Deposit(&_Escrow.TransactOpts, validator, amt)
}

// ReturnDeposit is a paid mutator transaction binding the contract method 0x05da2820.
//
// Solidity: function returnDeposit(address recipient, uint256 amt, uint256 fee, string nonce) returns()
func (_Escrow *EscrowTransactor) ReturnDeposit(opts *bind.TransactOpts, recipient common.Address, amt *big.Int, fee *big.Int, nonce string) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "returnDeposit", recipient, amt, fee, nonce)
}

// ReturnDeposit is a paid mutator transaction binding the contract method 0x05da2820.
//
// Solidity: function returnDeposit(address recipient, uint256 amt, uint256 fee, string nonce) returns()
func (_Escrow *EscrowSession) ReturnDeposit(recipient common.Address, amt *big.Int, fee *big.Int, nonce string) (*types.Transaction, error) {
	return _Escrow.Contract.ReturnDeposit(&_Escrow.TransactOpts, recipient, amt, fee, nonce)
}

// ReturnDeposit is a paid mutator transaction binding the contract method 0x05da2820.
//
// Solidity: function returnDeposit(address recipient, uint256 amt, uint256 fee, string nonce) returns()
func (_Escrow *EscrowTransactorSession) ReturnDeposit(recipient common.Address, amt *big.Int, fee *big.Int, nonce string) (*types.Transaction, error) {
	return _Escrow.Contract.ReturnDeposit(&_Escrow.TransactOpts, recipient, amt, fee, nonce)
}

// EscrowDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Escrow contract.
type EscrowDepositIterator struct {
	Event *EscrowDeposit // Event containing the contract specifics and raw log

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
func (it *EscrowDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowDeposit)
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
		it.Event = new(EscrowDeposit)
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
func (it *EscrowDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowDeposit represents a Deposit event raised by the Escrow contract.
type EscrowDeposit struct {
	Caller common.Address
	Target common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62.
//
// Solidity: event Deposit(address caller, address indexed target, uint256 amount)
func (_Escrow *EscrowFilterer) FilterDeposit(opts *bind.FilterOpts, target []common.Address) (*EscrowDepositIterator, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "Deposit", targetRule)
	if err != nil {
		return nil, err
	}
	return &EscrowDepositIterator{contract: _Escrow.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62.
//
// Solidity: event Deposit(address caller, address indexed target, uint256 amount)
func (_Escrow *EscrowFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *EscrowDeposit, target []common.Address) (event.Subscription, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "Deposit", targetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowDeposit)
				if err := _Escrow.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62.
//
// Solidity: event Deposit(address caller, address indexed target, uint256 amount)
func (_Escrow *EscrowFilterer) ParseDeposit(log types.Log) (*EscrowDeposit, error) {
	event := new(EscrowDeposit)
	if err := _Escrow.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EscrowWithdrawalIterator is returned from FilterWithdrawal and is used to iterate over the raw logs and unpacked data for Withdrawal events raised by the Escrow contract.
type EscrowWithdrawalIterator struct {
	Event *EscrowWithdrawal // Event containing the contract specifics and raw log

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
func (it *EscrowWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EscrowWithdrawal)
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
		it.Event = new(EscrowWithdrawal)
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
func (it *EscrowWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EscrowWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EscrowWithdrawal represents a Withdrawal event raised by the Escrow contract.
type EscrowWithdrawal struct {
	Receiver common.Address
	Caller   common.Address
	Amount   *big.Int
	Fee      *big.Int
	Nonce    string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterWithdrawal is a free log retrieval operation binding the contract event 0xfbe316db8c3cfdc314d06e8ab0b8baf61d03bd5aa62893c764ef1215a962d268.
//
// Solidity: event Withdrawal(address receiver, address indexed caller, uint256 amount, uint256 fee, string nonce)
func (_Escrow *EscrowFilterer) FilterWithdrawal(opts *bind.FilterOpts, caller []common.Address) (*EscrowWithdrawalIterator, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "Withdrawal", callerRule)
	if err != nil {
		return nil, err
	}
	return &EscrowWithdrawalIterator{contract: _Escrow.contract, event: "Withdrawal", logs: logs, sub: sub}, nil
}

// WatchWithdrawal is a free log subscription operation binding the contract event 0xfbe316db8c3cfdc314d06e8ab0b8baf61d03bd5aa62893c764ef1215a962d268.
//
// Solidity: event Withdrawal(address receiver, address indexed caller, uint256 amount, uint256 fee, string nonce)
func (_Escrow *EscrowFilterer) WatchWithdrawal(opts *bind.WatchOpts, sink chan<- *EscrowWithdrawal, caller []common.Address) (event.Subscription, error) {

	var callerRule []interface{}
	for _, callerItem := range caller {
		callerRule = append(callerRule, callerItem)
	}

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "Withdrawal", callerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EscrowWithdrawal)
				if err := _Escrow.contract.UnpackLog(event, "Withdrawal", log); err != nil {
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

// ParseWithdrawal is a log parse operation binding the contract event 0xfbe316db8c3cfdc314d06e8ab0b8baf61d03bd5aa62893c764ef1215a962d268.
//
// Solidity: event Withdrawal(address receiver, address indexed caller, uint256 amount, uint256 fee, string nonce)
func (_Escrow *EscrowFilterer) ParseWithdrawal(log types.Log) (*EscrowWithdrawal, error) {
	event := new(EscrowWithdrawal)
	if err := _Escrow.contract.UnpackLog(event, "Withdrawal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
