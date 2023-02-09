package kwil_client

import (
	"context"
	"fmt"
	"kwil/pkg/crypto/transactions"
	"kwil/pkg/databases"
	"kwil/pkg/databases/clean"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
	"strings"
)

func (c *Client) GetDatabaseSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error) {
	id := databases.GenerateSchemaName(owner, dbName)
	return c.GetDatabaseSchemaById(ctx, id)
}

func (c *Client) GetDatabaseSchemaById(ctx context.Context, id string) (*databases.Database[[]byte], error) {
	return c.Kwil.GetSchema(ctx, id)
}

func (c *Client) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*transactions.Response, error) {
	clean.Clean(db)

	// build tx
	tx, err := c.buildTx(ctx, db.Owner, transactions.DEPLOY_DATABASE, db)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction, type %d, err: %w", transactions.DEPLOY_DATABASE, err)
	}

	return c.Kwil.Broadcast(ctx, tx)
}

func (c *Client) DropDatabase(ctx context.Context, dbName string) (*transactions.Response, error) {
	owner := c.Config.Fund.GetAccountAddress()
	data := &databases.DatabaseIdentifier{
		Name:  dbName,
		Owner: owner,
	}

	// build tx
	tx, err := c.buildTx(ctx, owner, transactions.DROP_DATABASE, data)
	if err != nil {
		return nil, err
	}

	return c.Kwil.Broadcast(ctx, tx)
}

func (c *Client) ExecuteDatabase(ctx context.Context, dbName string, queryName string, queryInputs []*spec.KwilAny) (*transactions.Response, error) {
	owner := c.Config.Fund.GetAccountAddress()
	// create the dbid.  we will need this for the databases body
	dbId := databases.GenerateSchemaName(owner, dbName)

	qrs, err := c.Kwil.GetQueries(ctx, dbId)
	if err != nil {
		return nil, fmt.Errorf("failed to get executables: %w", err)
	}

	// get the query from the executables
	var query *executables.QuerySignature
	for _, q := range qrs {
		if strings.EqualFold(q.Name, queryName) {
			query = q
			break
		}
	}
	if query == nil {
		return nil, fmt.Errorf("query %s not found", queryName)
	}

	// check that each input is provided
	// this section kinda gross, should probably be refactored
	userIns := make([]*executables.UserInput, 0)
	for _, input := range query.Args {
		found := false
		for i := 0; i < len(queryInputs); i += 2 {
			if queryInputs[i].Value() == input.Name {
				found = true
				userIns = append(userIns, &executables.UserInput{
					Name:  input.Name,
					Value: queryInputs[i+1].Bytes(),
				})
				break
			}
		}
		if !found {
			return nil, fmt.Errorf(`input "%s" not provided`, input.Name)
		}
	}

	// create the databases body
	body := &executables.ExecutionBody{
		Database: dbId,
		Query:    query.Name,
		Inputs:   userIns,
	}

	// buildtx
	tx, err := c.buildTx(ctx, owner, transactions.EXECUTE_QUERY, body)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}

	// Broadcast
	res, err := c.Kwil.Broadcast(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to Broadcast transaction: %w", err)
	}
	return res, nil
}
