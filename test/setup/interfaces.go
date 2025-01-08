package setup

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
)

type KwilNode interface {
	PrivateKey() *crypto.Secp256k1PrivateKey
	PublicKey() *crypto.Secp256k1PublicKey
	IsValidator() bool
	Config() *config.Config
	JSONRPCClient(t *testing.T, ctx context.Context, usingGateway bool) (JSONRPCClient, error)
}

type JSONRPCClient interface {
	PrivateKey() crypto.PrivateKey
	PublicKey() crypto.PublicKey
	Ping(context.Context) (string, error)
}
