package client

import (
	"context"
	"fmt"
	commonpb "kwil/api/protobuf/common/v0/gen/go"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/x/types/execution"
	"kwil/x/utils/serialize"
)

func (c *Client) GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error) {
	res, err := c.txClt.GetExecutablesById(ctx, &txpb.GetExecutablesByIdRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get executable: %w", err)
	}

	return convertExecutables(res.Executables)
}

func convertExecutables(execs []*commonpb.Executable) ([]*execution.Executable, error) {
	convertedExecs := make([]*execution.Executable, len(execs))
	for i, exec := range execs {
		convExec, err := serialize.Convert[commonpb.Executable, execution.Executable](exec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert executable: %w", err)
		}

		convertedExecs[i] = convExec
	}

	return convertedExecs, nil
}
