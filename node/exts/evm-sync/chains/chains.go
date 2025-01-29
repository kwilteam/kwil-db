// package chains tracks the EVM chains that are supported by the node.
package chains

import (
	"fmt"
)

// ChainInfo is the information about a chain.
type ChainInfo struct {
	// Name is the name of the chain.
	// It is case-insensitive and unique.
	Name Chain
	// ID is the unique identifier of the chain.
	// e.g. Ethereum mainnet is 1.
	ID int
	// RequiredConfirmations is the number of confirmations required before an event is considered final.
	// For example, Ethereum mainnet requires 12 confirmations.
	RequiredConfirmations int64
}

func init() {
	err := registerChain(
		ChainInfo{
			Name:                  "ethereum",
			ID:                    1,
			RequiredConfirmations: 12,
		},
		ChainInfo{
			Name:                  "sepolia",
			ID:                    11155111,
			RequiredConfirmations: 12,
		},
	)
	if err != nil {
		panic(err)
	}
}

type Chain string

const (
	Ethereum Chain = "ethereum"
	Sepolia  Chain = "sepolia"
)

func (c Chain) Valid() error {
	switch c {
	case Ethereum, Sepolia:
		return nil
	default:
		return fmt.Errorf("invalid chain: %s", c)
	}
}

var registeredChains = map[Chain]ChainInfo{}

func registerChain(chains ...ChainInfo) error {
	for _, chain := range chains {
		if err := chain.Name.Valid(); err != nil {
			return err
		}

		_, ok := registeredChains[chain.Name]
		if ok {
			return fmt.Errorf("chain already registered: %s", chain.Name)
		}

		if chain.RequiredConfirmations < 1 {
			return fmt.Errorf("required confirmations must be >= 1: %s", chain.Name)
		}

		registeredChains[chain.Name] = chain
	}

	return nil
}

// GetChainInfo returns the chain information for the given chain.
func GetChainInfo(name Chain) (ChainInfo, bool) {
	chain, ok := registeredChains[name]
	return chain, ok
}
