package setup

import (
	"context"
	"crypto/rand"

	"github.com/kwilteam/kwil-db/core/client"
	cTypes "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
)

// jsonrpcGoDriver uses the Go client to interact with the kwil node
type jsonrpcGoDriver struct {
	cTypes.Client
	privateKey crypto.PrivateKey
	log        logFunc
}

var _ JSONRPCClient = (*jsonrpcGoDriver)(nil)

func newClient(ctx context.Context, endpoint string, usingGateway bool, l logFunc) (JSONRPCClient, error) {
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	if err != nil {
		return nil, err
	}
	secp256k1Priv := priv.(*crypto.Secp256k1PrivateKey)

	opts := &cTypes.Options{
		Signer: &auth.Secp256k1Signer{
			Secp256k1PrivateKey: *secp256k1Priv},
	}
	var cl cTypes.Client
	if usingGateway {
		cl, err = gatewayclient.NewClient(ctx, endpoint, &gatewayclient.GatewayOptions{
			Options: *opts,
		})
	} else {
		cl, err = client.NewClient(ctx, endpoint, opts)
	}
	if err != nil {
		return nil, err
	}

	return &jsonrpcGoDriver{
		privateKey: priv,
		Client:     cl,
		log:        l,
	}, nil
}

func (c *jsonrpcGoDriver) PrivateKey() crypto.PrivateKey {
	return c.privateKey
}

func (c *jsonrpcGoDriver) PublicKey() crypto.PublicKey {
	return c.privateKey.Public()
}
