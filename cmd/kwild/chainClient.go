package main

import (
	"fmt"
	"kwil/x/cfgx"
	chainClient "kwil/x/chain/client"
	ccDTO "kwil/x/chain/client/dto"
	ccService "kwil/x/chain/client/service"
)

// Builds the chain client from the meta config
func buildChainClient(cfg cfgx.Config) (chainClient.ChainClient, error) {
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

	requiredConfirmations, err := cfg.GetInt64("required-confirmations", 12)
	if err != nil {
		return nil, fmt.Errorf("error getting required confirmations from config: %d", err)
	}

	reconnectionInterval, err := cfg.GetInt64("reconnection-interval", 30)
	if err != nil {
		return nil, fmt.Errorf("error getting reconnectiong-interval from config: %d", err)
	}

	return ccService.NewChainClientExplicit(&ccDTO.Config{
		Endpoint:              providerEndpoint,
		ChainCode:             chainCode,
		ReconnectionInterval:  reconnectionInterval,
		RequiredConfirmations: requiredConfirmations,
	})
}
