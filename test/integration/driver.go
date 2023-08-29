package integration

import "github.com/kwilteam/kwil-db/test/specifications"

type KwilIntDriver interface {
	specifications.DatabaseDeployDsl
	specifications.DatabaseDropDsl
	specifications.ExecuteQueryDsl
	specifications.ExecuteCallDsl

	// TODO: enable when network ops are implemented
	//specifications.NetworkOpsDsl
}
