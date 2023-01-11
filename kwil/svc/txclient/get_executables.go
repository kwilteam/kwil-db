package txclient

import (
	"context"
	"fmt"
	"kwil/x/proto/commonpb"
	"kwil/x/proto/txpb"
	"kwil/x/types/databases"
	"kwil/x/types/execution"
	"kwil/x/utils/serialize"
)

func (c *client) GetExecutables(ctx context.Context, db *databases.DatabaseIdentifier) ([]*execution.Executable, error) {
	res, err := c.txs.GetExecutables(ctx, &txpb.GetExecutablesRequest{
		Owner:    db.Owner,
		Database: db.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get executables: %w", err)
	}

	return convertExecutables(res.Executables)
}

func (c *client) GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error) {
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
	for _, exec := range execs {
		convExec, err := serialize.Convert[commonpb.Executable, execution.Executable](exec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert executable: %w", err)
		}

		convertedExecs = append(convertedExecs, convExec)
	}

	return convertedExecs, nil
}
