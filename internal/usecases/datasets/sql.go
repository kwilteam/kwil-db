package datasets

import (
	"context"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/utils"
)

func (u *DatasetUseCase) UpdateBlockHeight(height int64) {
	u.engine.SetCurrentBlockHeight(height)
}

func (u *DatasetUseCase) BlockCommit(wal *utils.Wal, prevAppHash []byte) ([]byte, error) {
	// Go through all the datasets used in the block and commit the changes (shld we do this in a go routine?)
	ctx := context.Background()
	for dbid := range u.engine.ModifiedDBs {
		ds, err := u.engine.GetDataset(ctx, dbid)
		if err != nil {
			return nil, fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		// Store the changeset to persistent memory and Rollback changes
		cs, err := ds.BlockCommit()
		if err != nil {
			return nil, fmt.Errorf("block commit failed while checkpointing dataset %s Txs with error: %w", dbid, err)
		}
		u.engine.ModifiedDBs[dbid] = cs
		// d.ChangeSet = cs
		wal.UpdateMaxLineSz(len(cs))
		wal.WriteSync([]byte(dbid + "\n" + string(cs) + "\n"))
	}

	appHash := u.GenerateAppHash(prevAppHash)
	u.engine.ModifiedDBs = make(map[string][]byte)
	return appHash, nil
}

func (u *DatasetUseCase) ApplyChangesets(wal *utils.Wal) error {
	wal.Lock()
	defer wal.Unlock()

	scanner := wal.NewScanner()
	for scanner.Scan() {
		dbid := scanner.Text()
		cs := scanner.Text()

		ds, err := u.engine.GetDataset(context.Background(), dbid)
		if err != nil {
			return fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		err = ds.ApplyChangeset([]byte(cs))
		if err != nil {
			return fmt.Errorf("block commit failed while applying changeset for dataset %s with error: %w", dbid, err)
		}
	}

	wal.TruncateUnSafe()
	return nil
}

func (u *DatasetUseCase) GenerateAppHash(prevAppHash []byte) []byte {
	dbids := make([]string, 0, len(u.engine.ModifiedDBs))

	for dbid := range u.engine.ModifiedDBs {
		if (u.engine.ModifiedDBs[dbid]) == nil {
			continue
		}
		dbids = append(dbids, dbid)
	}

	if len(dbids) == 0 {
		return prevAppHash
	}

	sort.Strings(dbids)

	dbUpdates := ""
	for _, dbid := range dbids {
		dbUpdates += string(u.engine.ModifiedDBs[dbid])
	}

	dbHash := crypto.Sha256([]byte(dbUpdates))
	appHash := crypto.Sha256(append(prevAppHash, dbHash...))

	return appHash[:]
}
