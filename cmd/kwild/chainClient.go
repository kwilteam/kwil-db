package main

import (
	"fmt"
	"kwil/x/cfgx"
	"kwil/x/chain"
	ccDTO "kwil/x/chain/client/dto"
	ccService "kwil/x/chain/client/service"
	cp "kwil/x/chain/provider"
)

// Builds the chain client from the meta config
func buildChainClient(cfg cfgx.Config) (ccDTO.ChainClient, error) {
	config := cfg.Select("chain-client") // config specifically for chain client
	providerEndpoint := config.GetString("provider-endpoint", "")
	if providerEndpoint == "" {
		return nil, fmt.Errorf("chain-client.provider-endpoint must be a valid endpoint.  Received empty string / nil")
	}
	chainCode, err := config.GetInt64("chain-code", 0)
	if err != nil {
		return nil, fmt.Errorf("error getting chain-code from config: %d", err)
	}
	if chainCode == 0 {
		return nil, fmt.Errorf("invalid chain-code: %d", chainCode)
	}

	provider, err := cp.New(providerEndpoint, chain.ChainCode(chainCode))
	if err != nil {
		return nil, fmt.Errorf("error creating chain provider: %d", err)
	}

	requiredConfirmations, err := cfg.GetInt64("required-confirmations", 12)
	if err != nil {
		return nil, fmt.Errorf("error getting required confirmations from config: %d", err)
	}

	reconnectionInterval, err := cfg.GetInt64("reconnection-interval", 30)
	if err != nil {
		return nil, fmt.Errorf("error getting reconnectiong-interval from config: %d", err)
	}

	return ccService.NewChainClientExplicit(provider, &ccDTO.Config{
		ChainCode:             chainCode,
		ReconnectionInterval:  reconnectionInterval,
		RequiredConfirmations: requiredConfirmations,
	}), nil
}
