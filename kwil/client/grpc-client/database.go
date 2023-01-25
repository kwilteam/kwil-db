package grpc_client

import (
	"context"
	"fmt"
	"kwil/x/types/databases"
	"kwil/x/types/databases/clean"
	"kwil/x/types/execution"
	"kwil/x/types/transactions"
	"strings"
)

func (c *Client) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*transactions.Response, error) {
	clean.Clean(db)

	if !strings.EqualFold(db.Owner, c.Chain.GetConfig().Account) {
		return nil, fmt.Errorf("database owner must be the same as the current account.  Owner: %s, Account: %s", db.Owner, c.Chain.GetConfig().Account)
	}

	// build tx
	tx, err := c.BuildTransaction(ctx, transactions.DEPLOY_DATABASE, db, c.Chain.GetConfig().PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction, type %s, err: %w", transactions.DEPLOY_DATABASE, err)
	}

	return c.Txs.Broadcast(ctx, tx)
}

func (c *Client) GetDatabaseSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error) {
	return c.Txs.GetSchema(ctx, &databases.DatabaseIdentifier{
		Owner: owner,
		Name:  dbName,
	})
}

func (c *Client) DropDatabase(ctx context.Context, owner string, dbName string) (*transactions.Response, error) {
	data := &databases.DatabaseIdentifier{
		Name:  dbName,
		Owner: owner,
	}

	// build tx
	tx, err := c.BuildTransaction(ctx, transactions.DROP_DATABASE, data, c.Chain.GetConfig().PrivateKey)
	if err != nil {
		return nil, err
	}

	return c.Txs.Broadcast(ctx, tx)
}

func (c *Client) ExecuteDatabase(ctx context.Context, owner string, dbName string, queryName string, queryInputs []string) (*transactions.Response, error) {
	// create the dbid.  we will need this for the execution body
	dbId := databases.GenerateSchemaName(owner, dbName)

	executables, err := c.Txs.GetExecutablesById(ctx, dbId)
	if err != nil {
		return nil, fmt.Errorf("failed to get executables: %w", err)
	}

	// get the query from the executables
	var query *execution.Executable
	for _, executable := range executables {
		if strings.EqualFold(executable.Name, queryName) {
			query = executable
			break
		}
	}
	if query == nil {
		return nil, fmt.Errorf("query %s not found", queryName)
	}

	// check that each input is provided
	userIns := make([]*execution.UserInput, 0)
	for _, input := range query.UserInputs {
		found := false
		for i := 1; i < len(queryInputs); i += 2 {
			if queryInputs[i] == input.Name {
				found = true
				userIns = append(userIns, &execution.UserInput{
					Name:  input.Name,
					Value: queryInputs[i+1],
				})
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("input %s not provided", input.Name)
		}
	}

	// create the execution body
	body := &execution.ExecutionBody{
		Database: dbId,
		Query:    query.Name,
		Inputs:   userIns,
	}

	// buildtx
	tx, err := c.BuildTransaction(ctx, transactions.EXECUTE_QUERY, body, c.Chain.GetConfig().PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// broadcast
	res, err := c.Txs.Broadcast(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}
	return res, nil
}
