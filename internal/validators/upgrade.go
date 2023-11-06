package validators

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type upgradeAction int

const (
	upgradeActionNone          upgradeAction = iota // already at latest version
	upgradeActionInit                               // fresh DB, start at latest version
	upgradeActionRunMigrations                      // needs upgrade from older version
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
			return valStoreVersion, upgradeActionInit, nil
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
		return nil
	case upgradeActionInit:
		return vs.initTables(ctx)
	case upgradeActionRunMigrations:
		return vs.runMigrations(ctx, version)
	default:
		return fmt.Errorf("unknown upgrade action: %d", action)
	}
}

// runMigrations runs incremental db upgrades from current version to the latest version.
func (vs *validatorStore) runMigrations(ctx context.Context, version int) error {
	switch version {
	case 0:
		if err := vs.upgradeValidatorsDBfrom0To1(ctx); err != nil {
			return err
		}
		fallthrough
	case 1:
		if err := vs.upgradeValidatorsDBfrom1To2(ctx); err != nil {
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

// upgradeValidatorsDBfrom0To1 upgrades the validators db from version 0 to 1.
// Version 0: join_reqs table: [candidate, power]
// Version 1: join_reqs table: [candidate, power, expiryAt]
// "ALTER TABLE join_reqs ADD COLUMN expiresAt INTEGER;"
func (vs *validatorStore) upgradeValidatorsDBfrom0To1(ctx context.Context) error {
	vs.log.Info("Upgrading validators db from version 0 to 1")
	// Add v1 schema version table
	if err := vs.db.Execute(ctx, sqlInitVersionTableV1, nil); err != nil {
		return fmt.Errorf("failed to initialize schema version table: %w", err)
	}

	if err := vs.db.Execute(ctx, sqlAddJoinExpiryV1, nil); err != nil {
		return fmt.Errorf("failed to add expiresAt column to join_reqs table: %w", err)
	}

	const version = 1
	err := vs.db.Execute(ctx, sqlInitVersionRowV1, map[string]any{
		"$version": version,
	})
	if err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}

	return nil
}

// upgradeValidatorsDBfrom1To2 upgrades the validators db from version 1 to 2.
// Just create the removals table and bump the version.
func (vs *validatorStore) upgradeValidatorsDBfrom1To2(ctx context.Context) error {
	vs.log.Info("Upgrading validators db from version 1 to 2")
	if err := vs.db.Execute(ctx, sqlInitRemovalsTableV2, nil); err != nil {
		return fmt.Errorf("failed to create removals table: %w", err)
	}
	const version = 2
	return vs.updateCurrentVersion(ctx, version)
}
