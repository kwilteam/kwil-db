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
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_escrowToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"balance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amt\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"escrowToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"escrowedFunds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b506040516107b13803806107b1833981810160405281019061003191906100d4565b805f806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550506100ff565b5f80fd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6100a38261007a565b9050919050565b6100b381610099565b81146100bd575f80fd5b50565b5f815190506100ce816100aa565b92915050565b5f602082840312156100e9576100e8610076565b5b5f6100f6848285016100c0565b91505092915050565b6106a58061010c5f395ff3fe608060405260043610610049575f3560e01c80632fe319da1461004d578063b6b55f2514610077578063e3d670d714610093578063f0ec77fa146100cf578063fc7e286d146100f9575b5f80fd5b348015610058575f80fd5b50610061610135565b60405161006e919061039d565b60405180910390f35b610091600480360381019061008c91906103ed565b610158565b005b34801561009e575f80fd5b506100b960048036038101906100b49190610453565b6102c2565b6040516100c6919061048d565b60405180910390f35b3480156100da575f80fd5b506100e3610308565b6040516100f0919061048d565b60405180910390f35b348015610104575f80fd5b5061011f600480360381019061011a9190610453565b61030e565b60405161012c919061048d565b60405180910390f35b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5f8054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3330846040518463ffffffff1660e01b81526004016101b4939291906104b5565b6020604051808303815f875af11580156101d0573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906101f4919061051f565b610233576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161022a906105ca565b60405180910390fd5b8060015f3373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825461027f9190610615565b925050819055507fe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c33826040516102b7929190610648565b60405180910390a150565b5f60015f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b60025481565b6001602052805f5260405f205f915090505481565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f819050919050565b5f61036561036061035b84610323565b610342565b610323565b9050919050565b5f6103768261034b565b9050919050565b5f6103878261036c565b9050919050565b6103978161037d565b82525050565b5f6020820190506103b05f83018461038e565b92915050565b5f80fd5b5f819050919050565b6103cc816103ba565b81146103d6575f80fd5b50565b5f813590506103e7816103c3565b92915050565b5f60208284031215610402576104016103b6565b5b5f61040f848285016103d9565b91505092915050565b5f61042282610323565b9050919050565b61043281610418565b811461043c575f80fd5b50565b5f8135905061044d81610429565b92915050565b5f60208284031215610468576104676103b6565b5b5f6104758482850161043f565b91505092915050565b610487816103ba565b82525050565b5f6020820190506104a05f83018461047e565b92915050565b6104af81610418565b82525050565b5f6060820190506104c85f8301866104a6565b6104d560208301856104a6565b6104e2604083018461047e565b949350505050565b5f8115159050919050565b6104fe816104ea565b8114610508575f80fd5b50565b5f81519050610519816104f5565b92915050565b5f60208284031215610534576105336103b6565b5b5f6105418482850161050b565b91505092915050565b5f82825260208201905092915050565b7f4465706f736974206661696c65643a20746f6b656e20646964206e6f742073755f8201527f636365737366756c6c79207472616e7366657200000000000000000000000000602082015250565b5f6105b460338361054a565b91506105bf8261055a565b604082019050919050565b5f6020820190508181035f8301526105e1816105a8565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61061f826103ba565b915061062a836103ba565b9250828201905080821115610642576106416105e8565b5b92915050565b5f60408201905061065b5f8301856104a6565b610668602083018461047e565b939250505056fea2646970667358221220e6a6e9f40e22544e5f56b32074137f2e7bce310ca61538648d3a89bc84a673f064736f6c63430008160033",
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

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address wallet) view returns(uint256)
func (_Escrow *EscrowCaller) Balance(opts *bind.CallOpts, wallet common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "balance", wallet)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address wallet) view returns(uint256)
func (_Escrow *EscrowSession) Balance(wallet common.Address) (*big.Int, error) {
	return _Escrow.Contract.Balance(&_Escrow.CallOpts, wallet)
}

// Balance is a free data retrieval call binding the contract method 0xe3d670d7.
//
// Solidity: function balance(address wallet) view returns(uint256)
func (_Escrow *EscrowCallerSession) Balance(wallet common.Address) (*big.Int, error) {
	return _Escrow.Contract.Balance(&_Escrow.CallOpts, wallet)
}

// Deposits is a free data retrieval call binding the contract method 0xfc7e286d.
//
// Solidity: function deposits(address ) view returns(uint256)
func (_Escrow *EscrowCaller) Deposits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "deposits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Deposits is a free data retrieval call binding the contract method 0xfc7e286d.
//
// Solidity: function deposits(address ) view returns(uint256)
func (_Escrow *EscrowSession) Deposits(arg0 common.Address) (*big.Int, error) {
	return _Escrow.Contract.Deposits(&_Escrow.CallOpts, arg0)
}

// Deposits is a free data retrieval call binding the contract method 0xfc7e286d.
//
// Solidity: function deposits(address ) view returns(uint256)
func (_Escrow *EscrowCallerSession) Deposits(arg0 common.Address) (*big.Int, error) {
	return _Escrow.Contract.Deposits(&_Escrow.CallOpts, arg0)
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

// EscrowedFunds is a free data retrieval call binding the contract method 0xf0ec77fa.
//
// Solidity: function escrowedFunds() view returns(uint256)
func (_Escrow *EscrowCaller) EscrowedFunds(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Escrow.contract.Call(opts, &out, "escrowedFunds")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EscrowedFunds is a free data retrieval call binding the contract method 0xf0ec77fa.
//
// Solidity: function escrowedFunds() view returns(uint256)
func (_Escrow *EscrowSession) EscrowedFunds() (*big.Int, error) {
	return _Escrow.Contract.EscrowedFunds(&_Escrow.CallOpts)
}

// EscrowedFunds is a free data retrieval call binding the contract method 0xf0ec77fa.
//
// Solidity: function escrowedFunds() view returns(uint256)
func (_Escrow *EscrowCallerSession) EscrowedFunds() (*big.Int, error) {
	return _Escrow.Contract.EscrowedFunds(&_Escrow.CallOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amt) payable returns()
func (_Escrow *EscrowTransactor) Deposit(opts *bind.TransactOpts, amt *big.Int) (*types.Transaction, error) {
	return _Escrow.contract.Transact(opts, "deposit", amt)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amt) payable returns()
func (_Escrow *EscrowSession) Deposit(amt *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.Deposit(&_Escrow.TransactOpts, amt)
}

// Deposit is a paid mutator transaction binding the contract method 0xb6b55f25.
//
// Solidity: function deposit(uint256 amt) payable returns()
func (_Escrow *EscrowTransactorSession) Deposit(amt *big.Int) (*types.Transaction, error) {
	return _Escrow.Contract.Deposit(&_Escrow.TransactOpts, amt)
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
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address caller, uint256 amount)
func (_Escrow *EscrowFilterer) FilterDeposit(opts *bind.FilterOpts) (*EscrowDepositIterator, error) {

	logs, sub, err := _Escrow.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &EscrowDepositIterator{contract: _Escrow.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address caller, uint256 amount)
func (_Escrow *EscrowFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *EscrowDeposit) (event.Subscription, error) {

	logs, sub, err := _Escrow.contract.WatchLogs(opts, "Deposit")
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

// ParseDeposit is a log parse operation binding the contract event 0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c.
//
// Solidity: event Deposit(address caller, uint256 amount)
func (_Escrow *EscrowFilterer) ParseDeposit(log types.Log) (*EscrowDeposit, error) {
	event := new(EscrowDeposit)
	if err := _Escrow.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
