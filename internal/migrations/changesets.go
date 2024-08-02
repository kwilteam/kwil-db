package migrations

import (
	"context"
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
		app.Service.Logger.Info("insert changeset chunk", log.Int("height", migration.Height.Int64()), log.Int("chunkIndex", migration.ChunkIdx.Int64()))
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

		var currentHeight int64
		// get the last changeset height applied
		lastChangeset, err := getLastChangeset(ctx, tx)
		if err != nil {
			return err
		}

		if lastChangeset == -1 {
			currentHeight = migrationCfg.StartHeight
		} else {
			currentHeight = lastChangeset + 1
		}

		// Apply the changesets in order of block height
		for {
			// If the current height is greater than the migration end height, break
			if currentHeight >= migrationCfg.EndHeight {
				app.Service.Logger.Info("changeset migration completed", log.Int("height", currentHeight))
				return nil
			}

			// Check if all chunks have been received for the current height
			rcvd, err := allChunksReceived(ctx, tx, currentHeight)
			if err != nil {
				return err
			}

			if !rcvd {
				app.Service.Logger.Info("waiting for all chunks to be received", log.Int("height", currentHeight))
				break // all chunks are not received, wait for them
			}

			// retrieve the complete changeset for the current height
			cs, err := getChangesets(ctx, tx, currentHeight)
			if err != nil {
				return err
			}

			blockChangesets := &BlockChangesets{}
			err = blockChangesets.UnmarshalBinary(cs)
			if err != nil {
				return err
			}

			// Apply the changeset to the database
			changesetGroup := &pg.ChangesetGroup{
				Changesets: blockChangesets.Changesets,
			}
			if err = changesetGroup.ApplyChangesets(ctx, tx); err != nil {
				app.Service.Logger.Error("failed to apply changesets", log.Int("height", currentHeight), log.String("error", err.Error()))
				return err
			}

			// Apply the spends to the database
			for _, spend := range blockChangesets.Spends {
				if err = spend.ApplySpend(ctx, tx); err != nil {
					app.Service.Logger.Warn("failed to apply spend", log.Int("height", currentHeight), log.String("error", err.Error()))
				}
			}

			app.Service.Logger.Info("Applied changesets", log.Int("height", currentHeight), log.Int("size", len(cs)))

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
