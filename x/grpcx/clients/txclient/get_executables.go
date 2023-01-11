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

	convertedExecs := make([]*execution.Executable, len(res.Executables))
	for _, exec := range res.Executables {
		convExec, err := serialize.Convert[commonpb.Executable, execution.Executable](exec)
		if err != nil {
			return nil, fmt.Errorf("failed to convert executable: %w", err)
		}

		convertedExecs = append(convertedExecs, convExec)
	}

	return convertedExecs, nil
}
