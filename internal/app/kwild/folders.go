// Package kwild defines shared values pertaining to kwild's operation.
package kwild

// Top-level directory structure for the Server's systems. These are not user
// configurable.
const (
	ABCIDirName          = "abci" // cometBFT node's root folder
	ABCIInfoSubDirName   = "info" // e.g. abci/info for kv state data
	ApplicationDirName   = "application"
	ReceivedSnapsDirName = "rcvdSnaps"
	SigningDirName       = "signing"

	// Note that the sqlLite file path is user-configurable e.g. "data/kwil.db"
)
