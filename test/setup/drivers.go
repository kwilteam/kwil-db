package setup

import (
	"context"
)

type ClientDriver string

const (
	Go  ClientDriver = "go"
	CLI ClientDriver = "cli"
)

type Client interface {
	JSONRPCClient
}

type newClientFunc func(ctx context.Context, endpoint string, usingGateway bool, log logFunc) (Client, error)

func getNewClientFn(driver ClientDriver) newClientFunc {
	switch driver {
	case Go:
		return newClient
	case CLI:
		panic("CLI driver not implemented")
	default:
		panic("unknown driver")
	}
}

type logFunc func(string, ...any)
