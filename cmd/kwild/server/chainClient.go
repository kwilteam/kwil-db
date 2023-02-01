package server

import (
	"fmt"
	chainClient "kwil/pkg/chain/client"
	ccDTO "kwil/pkg/chain/client/dto"
	ccService "kwil/pkg/chain/client/service"
	"kwil/x/cfgx"
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

	requiredConfirmations, err := config.GetInt64("required-confirmations", 12)
	if err != nil {
		return nil, fmt.Errorf("error getting required confirmations from config: %d", err)
	}

	reconnectionInterval, err := config.GetInt64("reconnection-interval", 30)
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
