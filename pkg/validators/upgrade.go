package validators

import (
	"context"
	"fmt"
)

type upgradeAction int

const (
	upgradeActionNone upgradeAction = iota
	upgradeActionLegacy
	upgradeActionRunMigrations
)

/*
	CheckVersion checks the current version of the validator store and
	decides whether to run any migrations.
*/

func (vs *validatorStore) CheckVersion(ctx context.Context) (int, upgradeAction, error) {
	// Check if schema version exists
	version, versionErr := vs.currentVersion(ctx)

	// Check if validators db exists
	_, valErr := vs.currentValidators(ctx)

	fmt.Println("version: ", version, "versionErr: ", versionErr, "valStoreVersion: ", valStoreVersion)
	fmt.Println("valErr: ", valErr, "valStoreVersion: ", valStoreVersion)

	if versionErr != nil && valErr != nil {
		// Fresh db, do regular init
		return 1, upgradeActionNone, nil
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
	fmt.Println("version: ", version, "action: ", action, "err: ", err)
	if err != nil {
		return err
	}

	switch action {
	case upgradeActionNone:
		return vs.initTables(ctx)
	case upgradeActionLegacy:
	case upgradeActionRunMigrations:
		return vs.runMigrations(ctx, version)
	default:
		return fmt.Errorf("unknown upgrade action: %d", action)
	}
	return nil
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
	return nil
}
