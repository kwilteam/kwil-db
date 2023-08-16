package common

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/client"
)

const (
	// WithoutPrivateKey is a flag that can be passed to DialClient to indicate that the client should not use the private key in the config
	WithoutPrivateKey uint8 = 1 << iota
)

type RoundTripper func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error

func DialClient(ctx context.Context, flags uint8, fn RoundTripper) error {
	conf, err := config.LoadCliConfig()
	if err != nil {
		return err
	}

	options := []client.ClientOpt{}

	// We were previously mixing up the eth rpc url with the cometBFT RPC url.  Do we need to set it here?

	if flags&WithoutPrivateKey == 0 {
		// this means it needs to use the private key
		if conf.PrivateKey == nil {
			return fmt.Errorf("private key not provided")
		}

		options = append(options, client.WithSigner(conf.PrivateKey))
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

	pong, err := clt.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping provider: %w", err)
	}

	if pong != "pong" {
		return fmt.Errorf("unexpected ping response: %s", pong)
	}

	return fn(ctx, clt, conf)
}
