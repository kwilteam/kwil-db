package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/accounts"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/clean"
	"github.com/kwilteam/kwil-db/pkg/databases/convert"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

func (c *KwilClient) GetSchema(ctx context.Context, owner, name string) (*databases.Database[*spec.KwilAny], error) {
	return c.GetSchemaById(ctx, databases.GenerateSchemaId(owner, name))
}

func (c *KwilClient) GetSchemaById(ctx context.Context, id string) (*databases.Database[*spec.KwilAny], error) {
	byteDB, err := c.grpc.GetSchema(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema from provider: %w", err)
	}

	return convert.Bytes.DatabaseToKwilAny(byteDB)
}

func (c *KwilClient) DeployDatabase(ctx context.Context, db *databases.Database[[]byte], privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
	clean.Clean(db)

	// build tx
	tx, err := c.buildTx(ctx, accounts.DEPLOY_DATABASE, db, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction, type %d, err: %w", accounts.DEPLOY_DATABASE, err)
	}

	return c.grpc.Broadcast(ctx, tx)
}

func (c *KwilClient) DropDatabase(ctx context.Context, dbName string, privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
	owner, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	data := &databases.DatabaseIdentifier{
		Name:  dbName,
		Owner: owner,
	}

	// build tx
	tx, err := c.buildTx(ctx, accounts.DROP_DATABASE, data, privateKey)
	if err != nil {
		return nil, err
	}

	return c.grpc.Broadcast(ctx, tx)
}
