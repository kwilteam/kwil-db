package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/accounts"
	"kwil/pkg/crypto"
	"kwil/pkg/databases"
	"kwil/pkg/databases/clean"
	"kwil/pkg/databases/convert"
	"kwil/pkg/databases/spec"
)

func (c *client) GetSchema(ctx context.Context, owner, name string) (*databases.Database[*spec.KwilAny], error) {
	return c.GetSchemaById(ctx, databases.GenerateSchemaId(owner, name))
}

func (c *client) GetSchemaById(ctx context.Context, id string) (*databases.Database[*spec.KwilAny], error) {
	byteDB, err := c.grpc.GetSchema(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema from provider: %w", err)
	}

	return convert.Bytes.DatabaseToKwilAny(byteDB)
}

func (c *client) DeployDatabase(ctx context.Context, db *databases.Database[[]byte], privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
	clean.Clean(db)

	// build tx
	tx, err := c.buildTx(ctx, accounts.DEPLOY_DATABASE, db, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction, type %d, err: %w", accounts.DEPLOY_DATABASE, err)
	}

	return c.grpc.Broadcast(ctx, tx)
}

func (c *client) DropDatabase(ctx context.Context, dbName string, privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
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
