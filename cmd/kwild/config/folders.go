package config

// Top-level directory structure for the Server's systems. These are not user
// configurable.
const (
	ABCIDirName          = "abci" // cometBFT node's root folder
	ABCIInfoSubDirName   = "info" // e.g. abci/info for kv state data
	ApplicationDirName   = "application"
	ReceivedSnapsDirName = "rcvdSnaps"
	SigningDirName       = "signing"

	ConfigFileName     = "config.toml"
	PrivateKeyFileName = "private_key"

	// Note that the sqlLite file path is user-configurable e.g. "data/kwil.db"
)
