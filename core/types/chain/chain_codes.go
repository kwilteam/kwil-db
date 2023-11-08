package chain

import "math/big"

type ChainCode int // this is used to indicate the chain ID.
// I am using this instead of typically ChainIDs beecause we will want to support non-EVM chains (which have no chain ID)

const (
	LOCAL ChainCode = iota
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
	return "Local"
}

func (c *ChainCode) Int32() int32 {
	return int32(*c)
}
