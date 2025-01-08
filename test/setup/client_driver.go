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

// clientDriver uses the Go client to interact with the kwil node
type clientDriver struct {
	privateKey crypto.PrivateKey
	c          cTypes.Client
	log        logFunc
}

var _ JSONRPCClient = (*clientDriver)(nil)

func newClient(ctx context.Context, endpoint string, usingGateway bool, l logFunc) (Client, error) {
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

	return &clientDriver{
		privateKey: priv,
		c:          cl,
		log:        l,
	}, nil
}

func (c *clientDriver) PrivateKey() crypto.PrivateKey {
	return c.privateKey
}

func (c *clientDriver) PublicKey() crypto.PublicKey {
	return c.privateKey.Public()
}

func (c *clientDriver) Ping(ctx context.Context) (string, error) {
	return c.c.Ping(ctx)
}
