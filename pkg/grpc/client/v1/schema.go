package client

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*schema.Schema, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	ds, err := serialize.Convert[txpb.Dataset, schema.Schema](res.Dataset)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
