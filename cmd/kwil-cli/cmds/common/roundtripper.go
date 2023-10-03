package common

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/log"
)

const (
	// WithoutPrivateKey is a flag that can be passed to DialClient to indicate that the client should not use the private key in the config
	// this is a weird flag
	WithoutPrivateKey uint8 = 1 << iota
)

type RoundTripper func(ctx context.Context, client *client.Client, conf *config.KwilCliConfig) error

func DialClient(ctx context.Context, flags uint8, fn RoundTripper) error {
	conf, err := config.LoadCliConfig()
	if err != nil {
		return err
	}

	options := []client.Option{
		client.WithLogger(log.New(log.Config{
			Level: "error", // TODO: the log package should change this to take an enum instead of a string
		})),
		client.WithTLSCert(conf.TLSCertFile),
	}

	if flags&WithoutPrivateKey == 0 {
		// this means it needs to use the private key
		if conf.PrivateKey == nil {
			return fmt.Errorf("private key not provided")
		}

		signer := auth.EthPersonalSigner{Secp256k1PrivateKey: *conf.PrivateKey}
		options = append(options, client.WithSigner(&signer))
	}

	if conf.GrpcURL == "" {
		// the grpc url is required
		// this is somewhat redundant since the config marks it as required, but in case the config is changed
		return fmt.Errorf("kwil grpc url is required")
	}

	clt, err := client.Dial(conf.GrpcURL, options...)
	if err != nil {
		return err
	}
	defer clt.Close()

	pong, err := clt.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping provider: %w", err)
	}

	if pong != "pong" {
		return fmt.Errorf("unexpected ping response: %s", pong)
	}

	return fn(ctx, clt, conf)
}
