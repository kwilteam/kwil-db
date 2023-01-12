package txclient

import (
	"context"
	"fmt"
	"kwil/x/proto/commonpb"
	"kwil/x/proto/txpb"
	"kwil/x/types/execution"
	"kwil/x/utils/serialize"
)

func (c *client) GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error) {
	fmt.Println("id: ", id)
	res, err := c.txs.GetExecutablesById(ctx, &txpb.GetExecutablesByIdRequest{
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
