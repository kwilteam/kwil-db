package signersvc

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kwilteam/kwil-db/node/services/erc20signersvc/abigen"
)

// Multicall https://github.com/mds1/multicall
// https://etherscan.io/address/0xcA11bde05977b3631167028862bE2a173976CA11

const (
	EthNetworkMainnet = 1
	EthNetworkSepolia = 11155111
)

var (
	AddressMulticall3 = common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")
)

var deployedAtMap = map[uint64]uint64{
	EthNetworkMainnet: 14353601, // https://etherscan.io/tx/0x00d9fcb7848f6f6b0aae4fb709c133d69262b902156c85a473ef23faa60760bd
	EthNetworkSepolia: 751532,   // https://sepolia.etherscan.io/tx/0x6313b2cee1ddd9a77a8a1edf93495a9eb3c51a4d85479f4f8fec0090ad82596b
}

func IsMulticall3Deployed(chainID uint64, blockNumber *big.Int) bool {
	deployedAt, exists := deployedAtMap[chainID]
	if !exists {
		return false
	}

	if blockNumber == nil {
		return true
	}

	return deployedAt < blockNumber.Uint64()
}

// Aggregate3 aggregates multicall result.
// based on https://github.com/RSS3-Network/Node/blob/947b387f11857144c48250dd95804b5069731153/provider/ethereum/contract/multicall3/contract.go
func Aggregate3(ctx context.Context, chainID uint64, calls []abigen.Multicall3Call3,
	blockNumber *big.Int, contractBackend bind.ContractCaller) ([]*abigen.Multicall3Result, error) {
	if !IsMulticall3Deployed(chainID, blockNumber) {
		return nil, fmt.Errorf("multicall3 is not deployed on chainID %d yet", chainID)
	}

	abi, err := abigen.Multicall3MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("load abi: %w", err)
	}

	callData, err := abi.Pack("aggregate3", calls)
	if err != nil {
		return nil, fmt.Errorf("pack data: %w", err)
	}

	message := ethereum.CallMsg{
		To:   &AddressMulticall3,
		Data: callData,
	}

	results := make([]abigen.Multicall3Result, 0, len(calls))

	data, err := contractBackend.CallContract(ctx, message, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("call contract: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data in empty")
	}

	if err := abi.UnpackIntoInterface(&results, "aggregate3", data); err != nil {
		return nil, fmt.Errorf("unpack result: %w", err)
	}

	return ToSlicePtr(results), nil
}

func ToSlicePtr[T any](collection []T) []*T {
	result := make([]*T, len(collection))

	for i := range collection {
		result[i] = &collection[i]
	}
	return result
}
