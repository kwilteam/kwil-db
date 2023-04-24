package client

import (
	"context"
	"fmt"
	"kwil/pkg/databases/executables"
)

// this file handles responsiiblities for getting and storing databases and their queries

type dbi map[string]*executables.QuerySignature

// GetQuerySignature returns the query signature for the given query name
// it will download the dbi if it has not already been downloaded
func (c *KwilClient) GetQuerySignature(ctx context.Context, dbid, queryName string) (*executables.QuerySignature, error) {
	db, err := c.selectDB(ctx, dbid)
	if err != nil {
		return nil, err
	}

	qs, ok := db[queryName]
	if !ok {
		return nil, fmt.Errorf("query %s not found in database %s", queryName, dbid)
	}

	return qs, nil
}

// selectDB returns the dbi for the given id
// if the dbi has already been retrieved, it is returned
// otherwise, it is retrieved from the server and stored
func (c *KwilClient) selectDB(ctx context.Context, id string) (dbi, error) {
	db, ok := c.dbis[id]
	if ok {
		return db, nil
	}

	db = make(dbi)
	qrs, err := c.retrieveDBI(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, qr := range qrs {
		db[qr.Name] = qr
	}

	c.dbis[id] = db

	return db, nil
}

// retrieveDBI retrieves the dbi for the given id from the server
func (c *KwilClient) retrieveDBI(ctx context.Context, id string) ([]*executables.QuerySignature, error) {
	return c.grpc.GetQueries(ctx, id)
}
