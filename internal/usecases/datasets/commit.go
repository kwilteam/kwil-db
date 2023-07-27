package datasets

import (
	"context"
	"fmt"
)

func (u *DatasetUseCase) BlockCommit() error {
	// Go through all the datasets used in the block and commit the changes (shld we do this in a go routine?)
	ctx := context.Background()
	for dbid := range u.engine.ModifiedDBs {
		ds, err := u.engine.GetDataset(ctx, dbid) // Is the ds removed from the engine if it is dropped? flag if db is dropped
		if err != nil {
			return fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		// Get the savepoint on the db and commit
		err = ds.CheckpointAndCommit()
		if err != nil {
			return fmt.Errorf("block commit failed while checkpointing dataset %s Txs with error: %w", dbid, err)
		}
	}

	// Accounts need not be handled at the block boundaries

	// clear the modified list after commit
	u.engine.ModifiedDBs = make(map[string]bool)
	return nil
}

func (u *DatasetUseCase) UpdateBlockHeight(height int64) {
	u.engine.SetCurrentBlockHeight(height)
}
