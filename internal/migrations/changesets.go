package migrations

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

const (
	// "changeset_migration" is the event type used for replication of changesets from old chain to new chain during migration.
	ChangesetMigrationEventType = "changeset_migration"
)

var (
	migrationCfg *MigrationConfig

	ErrNoMoreChunksToRead = errors.New("no more chunks to read")
)

func init() {
	err := resolutions.RegisterResolution(ChangesetMigrationEventType, resolutions.ModAdd, ChangesetMigrationResolution)
	if err != nil {
		panic(err)
	}
}

// During network migration, the new node starts with the genesis state equal to the state of the old chain at the migration start height.
// The changeset migration system is used to replicate changesets from the old chain to the new chain in order to
// sync the state changes that occur on the old chain after the migration start height.
// These changesets are applied in the order of the block heights starting from the migration start height.
// Since the changesets can be very large and these changesets are to be sent through the voting system in the resolution body,
// and are constrained by the block sizes, the changesets are split into chunks and sent through the voting system.
// ChangesetMigration is the struct that represents the changeset migration chunk.
type ChangesetMigration struct {
	// Height is the block height the changeset belongs to.
	Height *big.Int

	// ChangesetIdx is the index of the changeset chunk in the block.
	// Indexes starts from 0.
	ChunkIdx *big.Int

	// TotalChunks is the total number of chunks in the changeset.
	TotalChunks *big.Int

	// Changeset is the serialized changeset chunk.
	Changeset []byte
}

// MarshalBinary marshals the ChangesetMigration into a binary format.
func (cs *ChangesetMigration) MarshalBinary() ([]byte, error) {
	return serialize.Encode(cs)
}

// UnmarshalBinary unmarshals the ChangesetMigration from a binary format.
func (cs *ChangesetMigration) UnmarshalBinary(data []byte) error {
	return serialize.Decode(data, cs)
}

// ChangesetMigrationResolution is the definition for changeset migration vote type in Kwil's voting system.
// ChangesetMigrationResolution is responsible for applying changesets from the old chain to the new chain.
// Once the changeset chunk is approved through the voting system, the chunk is added to the database.
// Once all the chunks for a particular height are received, the changeset is applied and the chunks are deleted.
// The changesets are applied in the order of block heights.
var ChangesetMigrationResolution = resolutions.ResolutionConfig{
	ConfirmationThreshold: big.NewRat(2, 3),
	ExpirationPeriod:      24 * 60 * 10, // 1 day in blocks
	ResolveFunc: func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error {

		// Extract the migration config from the app, executed only once
		if migrationCfg == nil {
			migrationCfg = &MigrationConfig{}
			cfg, ok := app.Service.ExtensionConfigs[ListenerName]
			if !ok {
				return errors.New("no migration config provided")
			}

			err := migrationCfg.ExtractConfig(cfg)
			if err != nil {
				return err
			}
		}

		var migration ChangesetMigration
		err := migration.UnmarshalBinary(resolution.Body)
		if err != nil {
			return err
		}

		// insert the changeset into the database
		app.Service.Logger.Debug("insert changeset chunk", log.Int("height", migration.Height.Int64()), log.Int("chunkIndex", migration.ChunkIdx.Int64()))
		tx, err := app.DB.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		// insert the changeset into the database
		err = migration.insertChangeset(ctx, tx)
		if err != nil {
			return err
		}

		// Commit the changeset chunk to the database, so that it is persisted
		// even if the resolution fails later
		if err = tx.Commit(ctx); err != nil {
			return err
		}

		tx, err = app.DB.BeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		var currentHeight int64
		// get the last changeset height applied
		lastChangeset, err := getLastChangeset(ctx, tx)
		if err != nil {
			return tx.Commit(ctx)
		}

		if lastChangeset == migrationCfg.EndHeight {
			return nil
		} else if lastChangeset == -1 {
			currentHeight = migrationCfg.StartHeight
		} else {
			currentHeight = lastChangeset + 1
		}

		// Apply the changesets in order of block height
		for {
			// If the current height is greater than the migration end height, break
			if currentHeight >= migrationCfg.EndHeight {
				app.Service.Logger.Info("changeset migration completed", log.Int("height", currentHeight))
				break
			}

			// Check if all chunks have been received for the current height, if not, break
			totalChunks, chunksReceived, err := getChangesetMetadata(ctx, tx, currentHeight)
			if err != nil {
				break
			}

			if totalChunks != chunksReceived {
				// app.Service.Logger.Info("waiting for all chunks to be received", log.Int("height", currentHeight))
				break
			}

			// Apply the changeset
			if err = applyChangeset(ctx, tx, currentHeight, totalChunks); err != nil {
				return err
			}

			app.Service.Logger.Info("Applied changesets", log.Int("height", currentHeight))

			// Delete the changeset after it has been applied
			if err = deleteChangesets(ctx, tx, currentHeight); err != nil {
				return err
			}

			// Increment the last changeset
			if err = setLastChangeset(ctx, tx, currentHeight); err != nil {
				return err
			}

			currentHeight += 1 // move to the next height
		}

		return tx.Commit(ctx)
	},
}

func applyChangeset(ctx context.Context, db sql.TxMaker, height int64, totalChunks int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	csReader := newChangesetReader(height, totalChunks)
	var data []byte

	var relations []*pg.Relation
	for {
		data, err = csReader.Read(ctx, tx, 5)
		if err != nil {
			if errors.Is(err, ErrNoMoreChunksToRead) {
				break
			}
			return err
		}

		csType := data[0]
		csSize := binary.LittleEndian.Uint32(data[1:5])

		data, err = csReader.Read(ctx, tx, int(csSize))
		if err != nil {
			return err
		}
		// Read the changeset type
		switch csType {
		case pg.RelationType:
			rel := &pg.Relation{}
			if err = rel.Deserialize(data); err != nil {
				return err
			}
			relations = append(relations, rel)

		case pg.ChangesetEntryType:
			ce := &pg.ChangesetEntry{}
			if err = ce.Deserialize(data); err != nil {
				return err
			}

			// apply the changeset entry
			if err = ce.ApplyChangesetEntry(ctx, tx, relations[ce.RelationIdx]); err != nil {
				return err
			}

		case pg.BlockSpendsType:
			bs := &BlockSpends{}
			if err = bs.Deserialize(data); err != nil {
				return err
			}

			// apply the block spends
			for _, spend := range bs.Spends {
				if err = spend.ApplySpend(ctx, tx); err != nil {
					return err
				}
			}

		default:
			return errors.New("unknown changeset type")
		}
	}
	return tx.Commit(ctx)
}

type changesetReader struct {
	height      int64
	chunkIdx    int64
	totalChunks int64
	data        []byte
}

func newChangesetReader(height int64, totalChunks int64) *changesetReader {
	return &changesetReader{
		height:      height,
		totalChunks: totalChunks,
	}
}

func (r *changesetReader) Read(ctx context.Context, tx sql.Executor, numBytesToRead int) ([]byte, error) {
	for len(r.data) < numBytesToRead {
		if r.chunkIdx >= r.totalChunks {
			return nil, ErrNoMoreChunksToRead
		}

		bts, err := getChangeset(ctx, tx, r.height, r.chunkIdx)
		if err != nil {
			return nil, err
		}

		r.data = append(r.data, bts...)
		r.chunkIdx += 1
	}

	data := r.data[:numBytesToRead]
	r.data = r.data[numBytesToRead:]

	return data, nil
}

// insertChangesetMigration inserts the changeset migration into the database.
// This inserts the changeset metadata for a particular height if it does not exist.
// It also inserts the changeset chunk into the database.
func (cm *ChangesetMigration) insertChangeset(ctx context.Context, db sql.TxMaker) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	//  Check if the changeset metadata entry exists, if not, create it
	if err := insertChangesetMetadata(ctx, tx, cm.Height.Int64(), int(cm.TotalChunks.Int64())); err != nil {
		return err
	}

	// check if this is not previously received
	var exists bool
	if exists, err = changesetChunkExists(ctx, tx, cm.Height.Int64(), int(cm.ChunkIdx.Int64())); err != nil {
		return err
	}

	if exists { // already received, ignore the changeset chunk
		return nil
	}

	// insert the changeset
	if err = insertChangesetChunk(ctx, tx, cm.Height.Int64(), int(cm.ChunkIdx.Int64()), cm.Changeset); err != nil {
		return err
	}

	// mark the chunk as received
	if _, err = tx.Execute(ctx, updateChangesetMetadataSQL, cm.Height); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
