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
)

// AbiMetaData contains all meta data concerning the Abi contract.
var AbiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_escrowToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"client\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountRequested\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountReturned\",\"type\":\"uint256\"}],\"name\":\"DepositReturned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message2\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"int64\",\"name\":\"myint\",\"type\":\"int64\"}],\"name\":\"GoodbyeWorld\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"message2\",\"type\":\"string\"}],\"name\":\"HelloWorld\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"RequestReturn\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"Amounts\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"escrowToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amt\",\"type\":\"uint256\"}],\"name\":\"requestReturn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"}],\"name\":\"requestReturnAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_fee\",\"type\":\"uint256\"}],\"name\":\"returnFunds\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AbiABI is the input ABI used to generate the binding from.
// Deprecated: Use AbiMetaData.ABI instead.
var AbiABI = AbiMetaData.ABI

// Abi is an auto generated Go binding around an Ethereum contract.
type Abi struct {
	AbiCaller     // Read-only binding to the contract
	AbiTransactor // Write-only binding to the contract
	AbiFilterer   // Log filterer for contract events
}

// AbiCaller is an auto generated read-only Go binding around an Ethereum contract.
type AbiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AbiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AbiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AbiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AbiSession struct {
	Contract     *Abi              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AbiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AbiCallerSession struct {
	Contract *AbiCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// AbiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AbiTransactorSession struct {
	Contract     *AbiTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AbiRaw is an auto generated low-level Go binding around an Ethereum contract.
type AbiRaw struct {
	Contract *Abi // Generic contract binding to access the raw methods on
}

// AbiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AbiCallerRaw struct {
	Contract *AbiCaller // Generic read-only contract binding to access the raw methods on
}

// AbiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AbiTransactorRaw struct {
	Contract *AbiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAbi creates a new instance of Abi, bound to a specific deployed contract.
func NewAbi(address common.Address, backend bind.ContractBackend) (*Abi, error) {
	contract, err := bindAbi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Abi{AbiCaller: AbiCaller{contract: contract}, AbiTransactor: AbiTransactor{contract: contract}, AbiFilterer: AbiFilterer{contract: contract}}, nil
}

// NewAbiCaller creates a new read-only instance of Abi, bound to a specific deployed contract.
func NewAbiCaller(address common.Address, caller bind.ContractCaller) (*AbiCaller, error) {
	contract, err := bindAbi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AbiCaller{contract: contract}, nil
}

// NewAbiTransactor creates a new write-only instance of Abi, bound to a specific deployed contract.
func NewAbiTransactor(address common.Address, transactor bind.ContractTransactor) (*AbiTransactor, error) {
	contract, err := bindAbi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AbiTransactor{contract: contract}, nil
}

// NewAbiFilterer creates a new log filterer instance of Abi, bound to a specific deployed contract.
func NewAbiFilterer(address common.Address, filterer bind.ContractFilterer) (*AbiFilterer, error) {
	contract, err := bindAbi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AbiFilterer{contract: contract}, nil
}

// bindAbi binds a generic wrapper to an already deployed contract.
func bindAbi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AbiABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Abi *AbiRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Abi.Contract.AbiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Abi *AbiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Abi.Contract.AbiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Abi *AbiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Abi.Contract.AbiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Abi *AbiCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Abi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Abi *AbiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Abi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Abi *AbiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Abi.Contract.contract.Transact(opts, method, params...)
}

// Amounts is a free data retrieval call binding the contract method 0xda6b4689.
//
// Solidity: function Amounts(address , address ) view returns(uint256)
func (_Abi *AbiCaller) Amounts(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "Amounts", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Amounts is a free data retrieval call binding the contract method 0xda6b4689.
//
// Solidity: function Amounts(address , address ) view returns(uint256)
func (_Abi *AbiSession) Amounts(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Abi.Contract.Amounts(&_Abi.CallOpts, arg0, arg1)
}

// Amounts is a free data retrieval call binding the contract method 0xda6b4689.
//
// Solidity: function Amounts(address , address ) view returns(uint256)
func (_Abi *AbiCallerSession) Amounts(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Abi.Contract.Amounts(&_Abi.CallOpts, arg0, arg1)
}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Abi *AbiCaller) EscrowToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Abi.contract.Call(opts, &out, "escrowToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Abi *AbiSession) EscrowToken() (common.Address, error) {
	return _Abi.Contract.EscrowToken(&_Abi.CallOpts)
}

// EscrowToken is a free data retrieval call binding the contract method 0x2fe319da.
//
// Solidity: function escrowToken() view returns(address)
func (_Abi *AbiCallerSession) EscrowToken() (common.Address, error) {
	return _Abi.Contract.EscrowToken(&_Abi.CallOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address _target, uint256 _amt) payable returns()
func (_Abi *AbiTransactor) Deposit(opts *bind.TransactOpts, _target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "deposit", _target, _amt)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address _target, uint256 _amt) payable returns()
func (_Abi *AbiSession) Deposit(_target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.Deposit(&_Abi.TransactOpts, _target, _amt)
}

// Deposit is a paid mutator transaction binding the contract method 0x47e7ef24.
//
// Solidity: function deposit(address _target, uint256 _amt) payable returns()
func (_Abi *AbiTransactorSession) Deposit(_target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.Deposit(&_Abi.TransactOpts, _target, _amt)
}

// RequestReturn is a paid mutator transaction binding the contract method 0xe212471f.
//
// Solidity: function requestReturn(address _target, uint256 _amt) returns()
func (_Abi *AbiTransactor) RequestReturn(opts *bind.TransactOpts, _target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "requestReturn", _target, _amt)
}

// RequestReturn is a paid mutator transaction binding the contract method 0xe212471f.
//
// Solidity: function requestReturn(address _target, uint256 _amt) returns()
func (_Abi *AbiSession) RequestReturn(_target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.RequestReturn(&_Abi.TransactOpts, _target, _amt)
}

// RequestReturn is a paid mutator transaction binding the contract method 0xe212471f.
//
// Solidity: function requestReturn(address _target, uint256 _amt) returns()
func (_Abi *AbiTransactorSession) RequestReturn(_target common.Address, _amt *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.RequestReturn(&_Abi.TransactOpts, _target, _amt)
}

// RequestReturnAll is a paid mutator transaction binding the contract method 0xd2199d0f.
//
// Solidity: function requestReturnAll(address _target) returns()
func (_Abi *AbiTransactor) RequestReturnAll(opts *bind.TransactOpts, _target common.Address) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "requestReturnAll", _target)
}

// RequestReturnAll is a paid mutator transaction binding the contract method 0xd2199d0f.
//
// Solidity: function requestReturnAll(address _target) returns()
func (_Abi *AbiSession) RequestReturnAll(_target common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RequestReturnAll(&_Abi.TransactOpts, _target)
}

// RequestReturnAll is a paid mutator transaction binding the contract method 0xd2199d0f.
//
// Solidity: function requestReturnAll(address _target) returns()
func (_Abi *AbiTransactorSession) RequestReturnAll(_target common.Address) (*types.Transaction, error) {
	return _Abi.Contract.RequestReturnAll(&_Abi.TransactOpts, _target)
}

// ReturnFunds is a paid mutator transaction binding the contract method 0x0ded499b.
//
// Solidity: function returnFunds(address _recipient, uint256 _amt, uint256 _fee) returns()
func (_Abi *AbiTransactor) ReturnFunds(opts *bind.TransactOpts, _recipient common.Address, _amt *big.Int, _fee *big.Int) (*types.Transaction, error) {
	return _Abi.contract.Transact(opts, "returnFunds", _recipient, _amt, _fee)
}

// ReturnFunds is a paid mutator transaction binding the contract method 0x0ded499b.
//
// Solidity: function returnFunds(address _recipient, uint256 _amt, uint256 _fee) returns()
func (_Abi *AbiSession) ReturnFunds(_recipient common.Address, _amt *big.Int, _fee *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.ReturnFunds(&_Abi.TransactOpts, _recipient, _amt, _fee)
}

// ReturnFunds is a paid mutator transaction binding the contract method 0x0ded499b.
//
// Solidity: function returnFunds(address _recipient, uint256 _amt, uint256 _fee) returns()
func (_Abi *AbiTransactorSession) ReturnFunds(_recipient common.Address, _amt *big.Int, _fee *big.Int) (*types.Transaction, error) {
	return _Abi.Contract.ReturnFunds(&_Abi.TransactOpts, _recipient, _amt, _fee)
}

// AbiDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Abi contract.
type AbiDepositIterator struct {
	Event *AbiDeposit // Event containing the contract specifics and raw log

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
func (it *AbiDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiDeposit)
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
		it.Event = new(AbiDeposit)
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
func (it *AbiDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiDeposit represents a Deposit event raised by the Abi contract.
type AbiDeposit struct {
	Caller common.Address
	Target common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62.
//
// Solidity: event Deposit(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) FilterDeposit(opts *bind.FilterOpts) (*AbiDepositIterator, error) {

	logs, sub, err := _Abi.contract.FilterLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return &AbiDepositIterator{contract: _Abi.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62.
//
// Solidity: event Deposit(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *AbiDeposit) (event.Subscription, error) {

	logs, sub, err := _Abi.contract.WatchLogs(opts, "Deposit")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiDeposit)
				if err := _Abi.contract.UnpackLog(event, "Deposit", log); err != nil {
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
// Solidity: event Deposit(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) ParseDeposit(log types.Log) (*AbiDeposit, error) {
	event := new(AbiDeposit)
	if err := _Abi.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiDepositReturnedIterator is returned from FilterDepositReturned and is used to iterate over the raw logs and unpacked data for DepositReturned events raised by the Abi contract.
type AbiDepositReturnedIterator struct {
	Event *AbiDepositReturned // Event containing the contract specifics and raw log

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
func (it *AbiDepositReturnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiDepositReturned)
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
		it.Event = new(AbiDepositReturned)
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
func (it *AbiDepositReturnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiDepositReturnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiDepositReturned represents a DepositReturned event raised by the Abi contract.
type AbiDepositReturned struct {
	Client          common.Address
	Target          common.Address
	AmountRequested *big.Int
	Fee             *big.Int
	AmountReturned  *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterDepositReturned is a free log retrieval operation binding the contract event 0x4609019aa7d7bb6817734dba73e107a725fd1edb218afbe7c5a3741bf9193fb9.
//
// Solidity: event DepositReturned(address client, address target, uint256 amountRequested, uint256 fee, uint256 amountReturned)
func (_Abi *AbiFilterer) FilterDepositReturned(opts *bind.FilterOpts) (*AbiDepositReturnedIterator, error) {

	logs, sub, err := _Abi.contract.FilterLogs(opts, "DepositReturned")
	if err != nil {
		return nil, err
	}
	return &AbiDepositReturnedIterator{contract: _Abi.contract, event: "DepositReturned", logs: logs, sub: sub}, nil
}

// WatchDepositReturned is a free log subscription operation binding the contract event 0x4609019aa7d7bb6817734dba73e107a725fd1edb218afbe7c5a3741bf9193fb9.
//
// Solidity: event DepositReturned(address client, address target, uint256 amountRequested, uint256 fee, uint256 amountReturned)
func (_Abi *AbiFilterer) WatchDepositReturned(opts *bind.WatchOpts, sink chan<- *AbiDepositReturned) (event.Subscription, error) {

	logs, sub, err := _Abi.contract.WatchLogs(opts, "DepositReturned")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiDepositReturned)
				if err := _Abi.contract.UnpackLog(event, "DepositReturned", log); err != nil {
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

// ParseDepositReturned is a log parse operation binding the contract event 0x4609019aa7d7bb6817734dba73e107a725fd1edb218afbe7c5a3741bf9193fb9.
//
// Solidity: event DepositReturned(address client, address target, uint256 amountRequested, uint256 fee, uint256 amountReturned)
func (_Abi *AbiFilterer) ParseDepositReturned(log types.Log) (*AbiDepositReturned, error) {
	event := new(AbiDepositReturned)
	if err := _Abi.contract.UnpackLog(event, "DepositReturned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiGoodbyeWorldIterator is returned from FilterGoodbyeWorld and is used to iterate over the raw logs and unpacked data for GoodbyeWorld events raised by the Abi contract.
type AbiGoodbyeWorldIterator struct {
	Event *AbiGoodbyeWorld // Event containing the contract specifics and raw log

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
func (it *AbiGoodbyeWorldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiGoodbyeWorld)
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
		it.Event = new(AbiGoodbyeWorld)
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
func (it *AbiGoodbyeWorldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiGoodbyeWorldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiGoodbyeWorld represents a GoodbyeWorld event raised by the Abi contract.
type AbiGoodbyeWorld struct {
	Message  string
	Message2 string
	Myint    int64
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterGoodbyeWorld is a free log retrieval operation binding the contract event 0xcfdda5facfe00c70a4d0d3ca711ffac2f660bb2a2e4624079b16ff612868a7d4.
//
// Solidity: event GoodbyeWorld(string message, string message2, int64 myint)
func (_Abi *AbiFilterer) FilterGoodbyeWorld(opts *bind.FilterOpts) (*AbiGoodbyeWorldIterator, error) {

	logs, sub, err := _Abi.contract.FilterLogs(opts, "GoodbyeWorld")
	if err != nil {
		return nil, err
	}
	return &AbiGoodbyeWorldIterator{contract: _Abi.contract, event: "GoodbyeWorld", logs: logs, sub: sub}, nil
}

// WatchGoodbyeWorld is a free log subscription operation binding the contract event 0xcfdda5facfe00c70a4d0d3ca711ffac2f660bb2a2e4624079b16ff612868a7d4.
//
// Solidity: event GoodbyeWorld(string message, string message2, int64 myint)
func (_Abi *AbiFilterer) WatchGoodbyeWorld(opts *bind.WatchOpts, sink chan<- *AbiGoodbyeWorld) (event.Subscription, error) {

	logs, sub, err := _Abi.contract.WatchLogs(opts, "GoodbyeWorld")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiGoodbyeWorld)
				if err := _Abi.contract.UnpackLog(event, "GoodbyeWorld", log); err != nil {
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

// ParseGoodbyeWorld is a log parse operation binding the contract event 0xcfdda5facfe00c70a4d0d3ca711ffac2f660bb2a2e4624079b16ff612868a7d4.
//
// Solidity: event GoodbyeWorld(string message, string message2, int64 myint)
func (_Abi *AbiFilterer) ParseGoodbyeWorld(log types.Log) (*AbiGoodbyeWorld, error) {
	event := new(AbiGoodbyeWorld)
	if err := _Abi.contract.UnpackLog(event, "GoodbyeWorld", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiHelloWorldIterator is returned from FilterHelloWorld and is used to iterate over the raw logs and unpacked data for HelloWorld events raised by the Abi contract.
type AbiHelloWorldIterator struct {
	Event *AbiHelloWorld // Event containing the contract specifics and raw log

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
func (it *AbiHelloWorldIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiHelloWorld)
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
		it.Event = new(AbiHelloWorld)
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
func (it *AbiHelloWorldIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiHelloWorldIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiHelloWorld represents a HelloWorld event raised by the Abi contract.
type AbiHelloWorld struct {
	Message  string
	Message2 string
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterHelloWorld is a free log retrieval operation binding the contract event 0xcc15048296af887a55840c391c10f70bd440808b98f569119bb9d39b92f1fd8c.
//
// Solidity: event HelloWorld(string message, string message2)
func (_Abi *AbiFilterer) FilterHelloWorld(opts *bind.FilterOpts) (*AbiHelloWorldIterator, error) {

	logs, sub, err := _Abi.contract.FilterLogs(opts, "HelloWorld")
	if err != nil {
		return nil, err
	}
	return &AbiHelloWorldIterator{contract: _Abi.contract, event: "HelloWorld", logs: logs, sub: sub}, nil
}

// WatchHelloWorld is a free log subscription operation binding the contract event 0xcc15048296af887a55840c391c10f70bd440808b98f569119bb9d39b92f1fd8c.
//
// Solidity: event HelloWorld(string message, string message2)
func (_Abi *AbiFilterer) WatchHelloWorld(opts *bind.WatchOpts, sink chan<- *AbiHelloWorld) (event.Subscription, error) {

	logs, sub, err := _Abi.contract.WatchLogs(opts, "HelloWorld")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiHelloWorld)
				if err := _Abi.contract.UnpackLog(event, "HelloWorld", log); err != nil {
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

// ParseHelloWorld is a log parse operation binding the contract event 0xcc15048296af887a55840c391c10f70bd440808b98f569119bb9d39b92f1fd8c.
//
// Solidity: event HelloWorld(string message, string message2)
func (_Abi *AbiFilterer) ParseHelloWorld(log types.Log) (*AbiHelloWorld, error) {
	event := new(AbiHelloWorld)
	if err := _Abi.contract.UnpackLog(event, "HelloWorld", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AbiRequestReturnIterator is returned from FilterRequestReturn and is used to iterate over the raw logs and unpacked data for RequestReturn events raised by the Abi contract.
type AbiRequestReturnIterator struct {
	Event *AbiRequestReturn // Event containing the contract specifics and raw log

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
func (it *AbiRequestReturnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AbiRequestReturn)
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
		it.Event = new(AbiRequestReturn)
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
func (it *AbiRequestReturnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AbiRequestReturnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AbiRequestReturn represents a RequestReturn event raised by the Abi contract.
type AbiRequestReturn struct {
	Caller common.Address
	Target common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterRequestReturn is a free log retrieval operation binding the contract event 0x776d05132488100ab9bb4599fa49d9da09d5f2d8b8bb71feaa021d6c44efebfe.
//
// Solidity: event RequestReturn(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) FilterRequestReturn(opts *bind.FilterOpts) (*AbiRequestReturnIterator, error) {

	logs, sub, err := _Abi.contract.FilterLogs(opts, "RequestReturn")
	if err != nil {
		return nil, err
	}
	return &AbiRequestReturnIterator{contract: _Abi.contract, event: "RequestReturn", logs: logs, sub: sub}, nil
}

// WatchRequestReturn is a free log subscription operation binding the contract event 0x776d05132488100ab9bb4599fa49d9da09d5f2d8b8bb71feaa021d6c44efebfe.
//
// Solidity: event RequestReturn(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) WatchRequestReturn(opts *bind.WatchOpts, sink chan<- *AbiRequestReturn) (event.Subscription, error) {

	logs, sub, err := _Abi.contract.WatchLogs(opts, "RequestReturn")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AbiRequestReturn)
				if err := _Abi.contract.UnpackLog(event, "RequestReturn", log); err != nil {
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

// ParseRequestReturn is a log parse operation binding the contract event 0x776d05132488100ab9bb4599fa49d9da09d5f2d8b8bb71feaa021d6c44efebfe.
//
// Solidity: event RequestReturn(address caller, address target, uint256 amount)
func (_Abi *AbiFilterer) ParseRequestReturn(log types.Log) (*AbiRequestReturn, error) {
	event := new(AbiRequestReturn)
	if err := _Abi.contract.UnpackLog(event, "RequestReturn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
