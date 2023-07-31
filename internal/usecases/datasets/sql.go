package datasets

import (
	"context"
	"fmt"
)

func (u *DatasetUseCase) UpdateBlockHeight(height int64) {
	u.engine.SetCurrentBlockHeight(height)
}

func (u *DatasetUseCase) BlockCommit() error {
	// Go through all the datasets used in the block and commit the changes (shld we do this in a go routine?)
	ctx := context.Background()
	for dbid := range u.engine.ModifiedDBs {
		ds, err := u.engine.GetDataset(ctx, dbid)
		if err != nil {
			return fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		// Store the changeset to persistent memory and Rollback changes
		err = ds.BlockCommit()
		if err != nil {
			return fmt.Errorf("block commit failed while checkpointing dataset %s Txs with error: %w", dbid, err)
		}
	}

	for dbid := range u.engine.ModifiedDBs {
		ds, err := u.engine.GetDataset(ctx, dbid)
		if err != nil {
			return fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		err = ds.ApplyChangeset()
		if err != nil {
			return fmt.Errorf("block commit failed while applying changeset for dataset %s with error: %w", dbid, err)
		}
	}

	// clear the modified list after commit
	u.engine.ModifiedDBs = make(map[string]bool)
	return nil
}
