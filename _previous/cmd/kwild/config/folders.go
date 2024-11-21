package config

import "path/filepath"

// Top-level directory structure for the Server's systems. These are not user
// configurable.
const (
	abciDirName    = "abci" // cometBFT node's root folder
	signingDirName = "signing"

	// ABCIInfoSubDirName is deprecated, only used to migrate old kv state
	// (meta) data into the main DB's kwild_chain schema (internal/abci/meta).
	abciInfoSubDirName = "info" // e.g. abci/info for kv state data

	configFileName    = "config.toml"
	migrationsDirName = "migrations"

	// receivedSnapshotsDirName is the directory where snapshots are received
	receivedSnapshotsDirName = "received_snapshots"
	// LocalSnapshots is the directory where snapshots taken by the local node are stored
	localSnapshotsDirName = "local_snapshots"

	genesisStateFileName = "genesis-state.sql.gz"
)

// GenesisStateFileName returns the genesis state file in the root directory.
func GenesisStateFileName(rootDir string) string {
	return filepath.Join(rootDir, genesisStateFileName)
}

// ReceivedSnapshotsDir returns the directory where snapshots are received
func ReceivedSnapshotsDir(rootDir string) string {
	return filepath.Join(rootDir, receivedSnapshotsDirName)
}

// LocalSnapshotsDir returns the directory where snapshots taken by the local node are stored
func LocalSnapshotsDir(rootDir string) string {
	return filepath.Join(rootDir, localSnapshotsDirName)
}

// ABCIDir returns the directory where the ABCI node's data is stored
func ABCIDir(rootDir string) string {
	return filepath.Join(rootDir, abciDirName)
}

// ABCIInfoDir returns the directory where the ABCI node's info is stored
func ABCIInfoDir(rootDir string) string {
	return filepath.Join(ABCIDir(rootDir), abciInfoSubDirName)
}

// SigningDir returns the directory where the ABCI node's signing keys are stored
func SigningDir(rootDir string) string {
	return filepath.Join(rootDir, signingDirName)
}

// ConfigFilePath returns the path to the config file
func ConfigFilePath(rootDir string) string {
	return filepath.Join(rootDir, configFileName)
}

// MigrationDir returns the directory where the node's migrations are stored
func MigrationDir(rootDir string) string {
	return filepath.Join(rootDir, migrationsDirName)
}
