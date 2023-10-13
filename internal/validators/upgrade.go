package validators

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type upgradeAction int

const (
	upgradeActionNone upgradeAction = iota
	upgradeActionRunMigrations
)

func upgradeActionString(action upgradeAction) string {
	switch action {
	case upgradeActionNone:
		return "none"
	case upgradeActionRunMigrations:
		return "run migrations"
	default:
		return "unknown"
	}
}

// checkVersion checks the current version of the validator store and decides
// whether to run any db migrations.
func (vs *validatorStore) checkVersion(ctx context.Context) (int, upgradeAction, error) {
	// Check if schema version exists
	version, versionErr := vs.currentVersion(ctx)
	// On error, infer that schema_version table doesn't exist (just assuming this
	// since we'd need to query sqlite_master table to be certain that was the error)
	if versionErr != nil {
		// Check if validators db exists (again only infers and this isn't robust because it could fail for other reasons)
		_, valErr := vs.currentValidators(ctx)
		if valErr != nil {
			// Fresh db, do regular initialization at valStoreVersion
			return valStoreVersion, upgradeActionNone, nil
		}
		if valErr == nil {
			// Legacy db without version tracking - version 0
			return 0, upgradeActionRunMigrations, nil
		}
	}

	if version == valStoreVersion {
		// DB on the latest version
		return version, upgradeActionNone, nil
	}
	if version < valStoreVersion {
		// DB on previous version, Run DB migrations
		return version, upgradeActionRunMigrations, nil
	}

	// Invalid DB version, return error
	return version, upgradeActionNone, fmt.Errorf("validator store version %d is higher than the supported version %d", version, valStoreVersion)
}

// databaseUpgrade runs the database upgrade based on the current version.
func (vs *validatorStore) initOrUpgradeDatabase(ctx context.Context) error {
	version, action, err := vs.checkVersion(ctx)
	if err != nil {
		return err
	}

	vs.log.Info("databaseUpgrade", zap.Int("version", version), zap.String("action", upgradeActionString(action)))

	switch action {
	case upgradeActionNone:
		return vs.initTables(ctx)
	case upgradeActionRunMigrations:
		return vs.runMigrations(ctx, version)
	default:
		vs.log.Error("unknown upgrade action", zap.Int("action", int(action)))
		return fmt.Errorf("unknown upgrade action: %d", action)
	}
}

// runMigrations runs incremental db upgrades from current version to the latest version.
func (vs *validatorStore) runMigrations(ctx context.Context, version int) error {
	switch version {
	case 0:
		if err := vs.upgradeValidatorsDB_0_1(ctx); err != nil {
			return err
		}
		fallthrough
	case valStoreVersion:
		vs.log.Info("databaseUpgrade: completed successfully")
		return nil
	default:
		vs.log.Error("unknown version", zap.Int("version", version))
		return fmt.Errorf("unknown version: %d", version)
	}
}

// upgradeValidatorsDB_0_1 upgrades the validators db from version 0 to 1.
// Version 0: join_reqs table: [candidate, power]
// Version 1: join_reqs table: [candidate, power, expiryAt]
// "ALTER TABLE join_reqs ADD COLUMN expiresAt INTEGER;"
func (vs *validatorStore) upgradeValidatorsDB_0_1(ctx context.Context) error {
	vs.log.Info("Upgrading validators db from version 0 to 1")
	// Add schema version table
	if err := vs.initSchemaVersion(ctx); err != nil {
		return err
	}

	if err := vs.db.Execute(ctx, sqlAddJoinExpiry, nil); err != nil {
		return fmt.Errorf("failed to add expiresAt column to join_reqs table: %w", err)
	}
	return nil
}
