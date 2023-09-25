package validators

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type upgradeAction int

const (
	upgradeActionNone upgradeAction = iota
	upgradeActionLegacy
	upgradeActionRunMigrations
)

func upgradeActionString(action upgradeAction) string {
	switch action {
	case upgradeActionNone:
		return "none"
	case upgradeActionLegacy:
		return "legacy"
	case upgradeActionRunMigrations:
		return "run migrations"
	default:
		return "unknown"
	}
}

/*
	CheckVersion checks the current version of the validator store and
	decides whether to run any migrations.
*/

func (vs *validatorStore) CheckVersion(ctx context.Context) (int, upgradeAction, error) {
	// Check if schema version exists
	version, versionErr := vs.currentVersion(ctx)

	// Check if validators db exists
	_, valErr := vs.currentValidators(ctx)

	if versionErr != nil && valErr != nil {
		// Fresh db, do regular init
		return valStoreVersion, upgradeActionNone, nil
	} else if versionErr != nil && valErr == nil {
		// Legacy db
		return 0, upgradeActionLegacy, nil
	} else if versionErr == nil && valErr == nil {
		// both tables exist
		if version == valStoreVersion {
			// Nothing to do
			return version, upgradeActionNone, nil
		} else if version < valStoreVersion {
			// Run DB migrations
			return version, upgradeActionRunMigrations, nil
		} else if version > valStoreVersion {
			// Error
			return version, upgradeActionNone, fmt.Errorf("validator store version %d is newer than the current version %d", version, valStoreVersion)
		}
	} else {
		// Error version == nil , valErr != nil, is it possible? Should we do regular init?
		return version, upgradeActionNone, fmt.Errorf("failed to check validator store version: %w", versionErr)
	}

	return version, upgradeActionNone, nil
}

func (vs *validatorStore) databaseUpgrade(ctx context.Context) error {
	version, action, err := vs.CheckVersion(ctx)
	vs.log.Info("databaseUpgrade", zap.String("version", fmt.Sprintf("%d", version)), zap.String("action", upgradeActionString(action)), zap.Error(err))

	if err != nil {
		return err
	}

	switch action {
	case upgradeActionNone:
		return vs.initTables(ctx)
	case upgradeActionLegacy:
		fallthrough
	case upgradeActionRunMigrations:
		return vs.runMigrations(ctx, version)
	default:
		return fmt.Errorf("unknown upgrade action: %d", action)
	}
}

func (vs *validatorStore) runMigrations(ctx context.Context, version int) error {
	switch version {
	case 0:
		if err := vs.upgradeValidatorsDB_0_1(ctx); err != nil {
			return err
		}
		// fallthrough
	default:
		return nil
	}
	return nil
}

/*
Version 0: join_reqs table: [candidate, power]
Version 1: join_reqs table: [candidate, power, expiryAt]
"ALTER TABLE join_reqs ADD COLUMN expiresAt INTEGER;"
*/
func (vs *validatorStore) upgradeValidatorsDB_0_1(ctx context.Context) error {
	vs.log.Info("Upgrading validators db from version 0 to 1")
	// Add schema version table
	if err := vs.initSchemaVersion(ctx); err != nil {
		return err
	}

	// Set schema version to 1
	if err := vs.setCurrentVersion(ctx, 1); err != nil {
		return err
	}

	if err := vs.db.Execute(ctx, "ALTER TABLE join_reqs ADD COLUMN expiresAt INTEGER;", nil); err != nil {
		return fmt.Errorf("failed to upgrade validators db from version 0 to 1: %w", err)
	}

	if err := vs.db.Execute(ctx, "UPDATE join_reqs SET expiresAt = -1;", nil); err != nil {
		return fmt.Errorf("failed to upgrade validators db from version 0 to 1: %w", err)
	}
	return nil
}
