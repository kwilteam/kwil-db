package setup

import (
	"context"

	client "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
)

type JSONRPCClient interface {
	client.Client
	PrivateKey() crypto.PrivateKey
	PublicKey() crypto.PublicKey
}

type ClientDriver string

const (
	Go  ClientDriver = "go"
	CLI ClientDriver = "cli"
)

type newClientFunc func(ctx context.Context, endpoint string, usingGateway bool, log logFunc) (JSONRPCClient, error)

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
