package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0"
	txpb "kwil/api/protobuf/tx/v0"
	"kwil/pkg/databases/executables"
	"kwil/pkg/utils/serialize"
)

func (c *Client) GetQueries(ctx context.Context, id string) ([]*executables.QuerySignature, error) {
	res, err := c.txClt.GetQueries(ctx, &txpb.GetQueriesRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get executable: %w", err)
	}

	return convertQuerySignatures(res.Queries)
}

func convertQuerySignatures(execs []*commonpb.QuerySignature) ([]*executables.QuerySignature, error) {
	convertedExecs := make([]*executables.QuerySignature, len(execs))
	for i, exec := range execs {
		convExec, err := serialize.Convert[commonpb.QuerySignature, executables.QuerySignature](exec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert executable: %w", err)
		}

		convertedExecs[i] = convExec
	}

	return convertedExecs, nil
}
