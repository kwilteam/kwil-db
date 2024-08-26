package operator

import "github.com/kwilteam/kwil-db/test/specifications"

// KwilOperatorDriver is the interface for a node operator.
type KwilOperatorDriver interface {
	specifications.ValidatorOpsDsl // TODO: split into ValidatorJoinDsl, ValidatorApproveDsl, ValidatorLeaveDsl
	specifications.ValidatorRemoveDsl
	specifications.PeersDsl
	specifications.MigrationOpsDsl
}
