package common

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	httpRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
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

	// right now the session/cookie check is only done in few APIs in KGW, not ideal
	// we won't all APIs until we have a rate limiter on KGW
	// NOTE: with KGW auth, we kind of always need cookie, hence user's wallet address
	// Store user's wallet address in config file? not sure

	if flags&WithoutPrivateKey == 0 {
		// this means it needs to use the private key
		if conf.PrivateKey == nil {
			return fmt.Errorf("private key not provided")
		}

		signer := auth.EthPersonalSigner{Key: *conf.PrivateKey}
		options = append(options, client.WithSigner(&signer, conf.ChainID))

		// try load kgw auth token from file, if exist
		// Kwild HTTP API doesn't care about KGW cookie, so not harm to load it
		addr, err := signer.Address()
		if err != nil {
			return fmt.Errorf("get address: %w", err)
		}
		kgwAuthInfo, err := LoadKGWAuthInfo(KGWAuthTokenFilePath(), addr)
		fmt.Printf("kgw auth info=========: %v\n", kgwAuthInfo)
		if err == nil && kgwAuthInfo != nil {
			// here create http client to config cookie
			// put cookie options in core/client/client.go seems not a good idea
			cookie := ConvertToHttpCookie(kgwAuthInfo.Cookie)
			hc, err := httpRPC.DialOptions(conf.GrpcURL, httpRPC.WithCookie(cookie))
			if err != nil {
				return err
			}
			options = append(options, client.WithRPCClient(hc))
		}
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

	return fn(ctx, clt, conf)
}
