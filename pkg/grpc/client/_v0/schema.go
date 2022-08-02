package client

import (
	"context"
	"fmt"
	commonpb "github.com/kwilteam/kwil-db/api/protobuf/common/v0"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

func (c *Client) GetSchema(ctx context.Context, id string) (*databases.Database[[]byte], error) {
	res, err := c.txClt.GetSchema(ctx, &txpb.GetSchemaRequest{
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
