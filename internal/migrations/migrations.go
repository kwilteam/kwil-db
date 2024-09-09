// package migrations implements a long-running migrations protocol for Kwil.
// This allows networks to upgrade to new networks over long periods of time,
// without any downtime.
//
// The process is as follows:
//
//  1. A network votes to create a new network. If enough votes are attained, the process is started.
//
//  2. Once the process is started, each validator should create a new node to run the new network, which will
//     connect to their current node. This new node will forward all changes from the old network to the new network.
//
//  3. The two networks will run in parallel until the old network reaches the scheduled shutdown block. At this point,
//     the new network will take over and the old network will be shut down.
//
// The old network cannot deploy databases, drop them, transfer balances, vote on any resolutions, or change their validator power.
//
// For more information on conflict resolution, see https://github.com/kwilteam/kwil-db/wiki/Long%E2%80%90Running-Network-Migrations
package migrations

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/voting"
)

// migrator instance responsible for managing zero downtime migrations.
var migrator *Migrator

func init() {
	migrator = &Migrator{}

	err := resolutions.RegisterResolution(voting.StartMigrationEventType, resolutions.ModAdd, MigrationResolution)
	if err != nil {
		panic(err)
	}
}

// MigrationDeclaration creates a new migration. It is used to agree on terms of a migration,
// and is voted on using Kwil's vote store.
type MigrationDeclaration struct {
	// ActivationPeriod is the amount of blocks before the migration is activated.
	// It starts after the migration is approved via the voting system.
	// The intention is to allow validators to prepare for the migration.
	ActivationPeriod uint64
	// Duration is the amount of blocks the migration will take to complete.
	Duration uint64
	// ChainID is the new chain ID that the network will migrate to.
	// A new chain ID should always be used for a new network, to avoid
	// cross-network replay attacks.
	ChainID string
	// Timestamp is the time the migration was created. It is set by the migration
	// creator. The primary purpose of it is to guarantee uniqueness of the serialized
	// MigrationDeclaration, since that is a requirement for the voting system.
	Timestamp string
}

// MarshalBinary marshals the MigrationDeclaration into a binary format.
func (md *MigrationDeclaration) MarshalBinary() ([]byte, error) {
	return serialize.Encode(md)
}

// UnmarshalBinary unmarshals the MigrationDeclaration from a binary format.
func (md *MigrationDeclaration) UnmarshalBinary(data []byte) error {
	return serialize.Decode(data, md)
}

// MigrationResolution is the definition for the network migration vote type in Kwil's
// voting system.
var MigrationResolution = resolutions.ResolutionConfig{
	ConfirmationThreshold: big.NewRat(2, 3),
	ExpirationPeriod:      100800, // 1 week
	ResolveFunc: func(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error {
		if migrator == nil || !migrator.initialized {
			return fmt.Errorf("migrator not initialized")
		}
		return migrator.startMigration(ctx, app, resolution, block)
	},
}

func (m *Migrator) startMigration(ctx context.Context, app *common.App, resolution *resolutions.Resolution, block *common.BlockContext) error {
	// check if the node is in migration mode already
	if block.ChainContext.MigrationParams != nil {
		app.Service.Logger.Warn("node is currently migrating from the old chain. Resubmit the migration proposal after the current migration is complete")
		return nil
	}

	// check if there is already an active migration
	alreadyHasMigration, err := migrationActive(ctx, app.DB)
	if err != nil {
		return err
	}

	if alreadyHasMigration {
		return fmt.Errorf("failed to start migration: only one migration can be active at a time")
	}

	mig := &MigrationDeclaration{}
	if err := mig.UnmarshalBinary(resolution.Body); err != nil {
		return err
	}

	activationPeriod := int64(mig.ActivationPeriod)
	dur := int64(mig.Duration)

	// the start height for the migration is whatever the height the migration
	// resolution passed + the activation period, which allows validators to prepare
	// for the migration. End height is the same, + the duration of the migration.

	active := &activeMigration{
		StartHeight: block.Height + activationPeriod,
		EndHeight:   block.Height + activationPeriod + dur,
		ChainID:     mig.ChainID,
	}

	err = createMigration(ctx, app.DB, active)
	if err != nil {
		return err
	}

	block.ChainContext.NetworkParameters.MigrationStatus = types.MigrationNotStarted
	m.activeMigration = active
	app.Service.Logger.Info("migration started", log.Int("start_height", active.StartHeight), log.Int("end_height", active.EndHeight))

	// Delete the pending migration resolutions from the resolutions table
	if err = voting.DeleteResolutionsByType(ctx, app.DB, []string{voting.StartMigrationEventType}); err != nil {
		return err
	}

	return nil
}
