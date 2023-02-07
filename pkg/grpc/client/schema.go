package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/kwil/common/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/pkg/types/databases"
	"kwil/pkg/utils/serialize"
)

func (c *Client) GetSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error) {
	res, err := c.txClt.GetSchema(ctx, &txpb.GetSchemaRequest{
		Owner: owner,
		Name:  dbName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	return convertDatabase(res.Database)
}

func (c *Client) GetSchemaById(ctx context.Context, id string) (*databases.Database[[]byte], error) {
	res, err := c.txClt.GetSchemaById(ctx, &txpb.GetSchemaByIdRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	return convertDatabase(res.Database)
}

func convertDatabase(db *commonpb.Database) (*databases.Database[[]byte], error) {
	// convert tables
	// convert response to database
	dbRes, err := serialize.Convert[commonpb.Database, databases.Database[[]byte]](db)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response: %w", err)
	}

	return dbRes, nil
}
