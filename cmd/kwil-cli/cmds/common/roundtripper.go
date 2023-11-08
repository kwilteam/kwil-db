package common

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
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

		signer := auth.EthPersonalSigner{Key: *conf.PrivateKey}
		options = append(options, client.WithSigner(&signer, conf.ChainID))
	}

	if conf.GrpcURL == "" {
		// the grpc url is required
		// this is somewhat redundant since the config marks it as required, but in case the config is changed
		return fmt.Errorf("kwil grpc url is required")
	}

	clt, err := client.Dial(ctx, conf.GrpcURL, options...)
	if err != nil {
		return err
	}
	defer clt.Close()

	// TODO:
	/*
		if conf.ChainRPCURL != "" {
			create a chainClient
		}
		get the token and escrow contract instances. not here, but in the correspondong method functions
	*/
	return fn(ctx, clt, conf)
}
