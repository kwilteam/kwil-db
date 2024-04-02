package config

// Top-level directory structure for the Server's systems. These are not user
// configurable.
const (
	ABCIDirName          = "abci" // cometBFT node's root folder
	ReceivedSnapsDirName = "rcvdSnaps"
	SigningDirName       = "signing"

	// ABCIInfoSubDirName is deprecated, only used to migrate old kv state
	// (meta) data into the main DB's kwild_chain schema (internal/abci/meta).
	ABCIInfoSubDirName = "info" // e.g. abci/info for kv state data

	ConfigFileName     = "config.toml"
	PrivateKeyFileName = "private_key"

	SnapshotDirName = "snapshots"

	// Note that the sqlLite file path is user-configurable e.g. "data/kwil.db"
)
