package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/accounts"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
)

func (c *KwilClient) ExecuteDatabase(ctx context.Context, dbOwner, dbName string, queryName string, queryInputs map[string]*spec.KwilAny, privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
	dbid := databases.GenerateSchemaId(dbOwner, dbName)

	return c.ExecuteDatabaseById(ctx, dbid, queryName, queryInputs, privateKey)
}

func (c *KwilClient) ExecuteDatabaseById(ctx context.Context, dbid string, queryName string, queryInputs map[string]*spec.KwilAny, privateKey *ecdsa.PrivateKey) (*accounts.Response, error) {
	qrs, err := c.GetQuerySignature(ctx, dbid, queryName)
	if err != nil {
		return nil, fmt.Errorf("failed to get query signature: %w", err)
	}

	userInputs := make([]*executables.UserInput, 0)
	for _, arg := range qrs.Args {
		input, ok := queryInputs[arg.Name]
		if !ok {
			return nil, fmt.Errorf(`required input "%s" not provided`, arg.Name)
		}
		userInputs = append(userInputs, &executables.UserInput{
			Name:  arg.Name,
			Value: input.Bytes(),
		})
	}

	// create the databases body
	body := &executables.ExecutionBody{
		Database: dbid,
		Query:    qrs.Name,
		Inputs:   userInputs,
	}

	// buildtx
	tx, err := c.buildTx(ctx, accounts.EXECUTE_QUERY, body, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Broadcast
	res, err := c.grpc.Broadcast(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to Broadcast transaction: %w", err)
	}
	return res, nil
}
