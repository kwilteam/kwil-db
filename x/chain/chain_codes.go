package chain

import "math/big"

type ChainCode int // this is used to indicate the chain ID.
// I am using this instead of typicaly ChainIDs beecause we will want to support non-EVM chains (which have no chain ID)

const (
	UNKNOWN_CHAIN ChainCode = iota
	ETHEREUM
	GOERLI
)

func (c ChainCode) ToChainId() *big.Int {
	switch c {
	case ETHEREUM:
		return big.NewInt(1)
	case GOERLI:
		return big.NewInt(5)
	}
	return big.NewInt(0)
}

func (c ChainCode) String() string {
	switch c {
	case ETHEREUM:
		return "Ethereum"
	case GOERLI:
		return "Goerli"
	}
	return "Unknown"
}
