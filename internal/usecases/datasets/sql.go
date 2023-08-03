package datasets

import (
	"context"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	gowal "github.com/tidwall/wal"
)

func (u *DatasetUseCase) UpdateBlockHeight(height int64) {
	u.engine.SetCurrentBlockHeight(height)
}

func BatchWrite(wal *gowal.Log, dbid string, cs []byte, idx uint64) error {
	batch := new(gowal.Batch)
	batch.Write(idx, []byte(dbid))
	batch.Write(idx+1, cs)
	return wal.WriteBatch(batch)
}

func (u *DatasetUseCase) BlockCommit(wal *gowal.Log, prevAppHash []byte) ([]byte, error) {
	// Go through all the datasets used in the block and retrieve the changesets, write them to the shadow wal
	ctx := context.Background()

	lastIdx, _ := wal.LastIndex()
	idx := lastIdx + 1

	for dbid := range u.engine.ModifiedDBs {
		ds, err := u.engine.GetDataset(ctx, dbid)
		if err != nil {
			return nil, fmt.Errorf("block commit failed while retrieving dataset %s with error: %w", dbid, err)
		}

		// write the changeset to shadowwal and Rollback changes on the sql wal
		cs, err := ds.BlockCommit()
		if err != nil {
			return nil, fmt.Errorf("block commit failed while checkpointing dataset %s Txs with error: %w", dbid, err)
		}
		u.engine.ModifiedDBs[dbid] = cs

		if cs != nil {
			err = BatchWrite(wal, dbid, cs, idx)
			if err != nil {
				return nil, fmt.Errorf("writing changesets to wal failed with err: %v", err)
			}
			idx += 2
		}
	}

	appHash := u.GenerateAppHash(prevAppHash)
	u.engine.ModifiedDBs = make(map[string][]byte) // Do this later? after the wal is checkpointed?

	return appHash, nil
}

func (u *DatasetUseCase) ApplyChangesets(wal *gowal.Log) error {

	firstIdx, _ := wal.FirstIndex()
	lastIdx, _ := wal.LastIndex()
	idx := firstIdx + 1

	if firstIdx == lastIdx {
		return nil
	}

	for idx <= lastIdx-1 {
		dbid, err := wal.Read(idx)
		if err != nil {
			return fmt.Errorf("reading from block wal failed at index %d with error: %w", idx, err)
		}
		ds, err := u.engine.GetDataset(context.Background(), string(dbid))
		if err != nil {
			return fmt.Errorf("dataset retrieval for %s failed with error: %w", dbid, err)
		}
		cs, err := wal.Read(idx + 1)
		if err != nil {
			return fmt.Errorf("reading from block wal failed at index %d with error: %w", idx+1, err)
		}
		err = ds.ApplyChangeset(cs)
		if err != nil {
			return fmt.Errorf("apply changesets for dataset %s failed with error: %w", dbid, err)
		}
		idx += 2
	}

	wal.TruncateBack(firstIdx)
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
