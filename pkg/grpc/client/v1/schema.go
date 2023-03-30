package client

import (
	"context"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/pkg/engine/models"
	"kwil/pkg/utils/serialize"
)

func (c *Client) GetSchema(ctx context.Context, dbid string) (*models.Dataset, error) {
	res, err := c.txClient.GetSchema(ctx, &txpb.GetSchemaRequest{
		Dbid: dbid,
	})
	if err != nil {
		return nil, err
	}

	ds, err := serialize.Convert[txpb.Dataset, models.Dataset](res.Dataset)
	if err != nil {
		return nil, err
	}

	return ds, nil
}
