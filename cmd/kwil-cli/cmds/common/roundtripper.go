package common

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
)

const (
	// WithoutServiceConfig is a flag that can be passed to DialClient to indicate that the client should not use the server's chain config
	WithoutServiceConfig uint8 = 1 << iota

	// WithoutProvider is a flag that can be passed to DialClient to indicate that the client should not establish a connection to the provider
	WithoutProvider

	// WithChainClient is a flag that can be passed to DialClient to indicate that the client should be configured to use a chain client.
	// If no ChainRPCURL is provided in the config, this will return an error
	WithChainClient

	// WithoutPrivateKey is a flag that can be passed to DialClient to indicate that the client should not use the private key in the config
	WithoutPrivateKey
)

type RoundTripper func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error

func DialClient(ctx context.Context, flags uint8, fn RoundTripper) error {
	conf, err := config.LoadCliConfig()
	if err != nil {
		return err
	}

	options := []client.ClientOpt{}

	if flags&WithoutProvider != 0 {
		options = append(options, client.WithoutProvider())
	}
	if flags&WithoutServiceConfig != 0 {
		options = append(options, client.WithoutServiceConfig())
	}
	if flags&WithChainClient != 0 {
		if conf.ClientChainRPCURL == "" {
			return fmt.Errorf("chain rpc url is required")
		}
		options = append(options, client.WithChainRpcUrl(conf.ClientChainRPCURL))
	}
	if flags&WithoutPrivateKey == 0 {
		// this means it needs to use the private key
		if conf.PrivateKey == nil {
			return fmt.Errorf("private key not provided")
		}

		options = append(options, client.WithPrivateKey(conf.PrivateKey))
	}

	if conf.GrpcURL == "" {
		// the grpc url is required
		// this is somewhat redundant since the config marks it as required, but in case the config is changed
		return fmt.Errorf("kwil grpc url is required")
	}

	clt, err := client.New(ctx, conf.GrpcURL,
		options...,
	)
	if err != nil {
		return err
	}

	if flags&WithoutProvider == 0 {
		pong, err := clt.Ping(ctx)
		if err != nil {
			return fmt.Errorf("failed to ping provider: %w", err)
		}

		if pong != "pong" {
			return fmt.Errorf("unexpected ping response: %s", pong)
		}
	}
	return fn(ctx, clt, conf)
}
