package setup

import (
	"context"
	"crypto/rand"

	client "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"
)

type JSONRPCClient interface {
	client.Client
	PrivateKey() crypto.PrivateKey
	PublicKey() crypto.PublicKey
	TxSuccess(ctx context.Context, txHash types.Hash) error
	Identifier() string
}

// ClientOptions allows a test to configure a client.
// They are all optional.
type ClientOptions struct {
	// PrivateKey is the private key to use for the client.
	PrivateKey crypto.PrivateKey
	// UsingKGW specifies whether to use the gateway client.
	UsingKGW bool
	// ChainID is the chain ID to use.
	ChainID string
}

func (c *ClientOptions) ensureDefaults() {
	if c.PrivateKey == nil {
		pk, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
		if err != nil {
			panic(err)
		}

		c.PrivateKey = pk
	}
	if c.ChainID == "" {
		c.ChainID = "kwil-test-chain"
	}
}

type ClientDriver string

var AllDrivers = []ClientDriver{Go, CLI}

const (
	Go  ClientDriver = "go"
	CLI ClientDriver = "cli"
)

func (d ClientDriver) String() string {
	return string(d)
}

type newClientFunc func(ctx context.Context, endpoint string, log logFunc, opts *ClientOptions) (JSONRPCClient, error)

func getNewClientFn(driver ClientDriver) newClientFunc {
	switch driver {
	case Go:
		return newClient
	case CLI:
		return newKwilCI
	default:
		panic("unknown driver")
	}
}

type logFunc func(string, ...any)
