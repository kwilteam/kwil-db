package adminsvc

import (
	"github.com/kwilteam/kwil-db/node/services/jsonrpc/openrpc"
)

var (
	SpecInfo = openrpc.Info{
		Title:       "Kwil DB admin service",
		Description: `The JSON-RPC admin service for Kwil DB.`,
		License: &openrpc.License{ // the spec file's license
			Name: "CC0-1.0",
			URL:  "https://creativecommons.org/publicdomain/zero/1.0/legalcode",
		},
		Version: "0.1.0",
	}
)
