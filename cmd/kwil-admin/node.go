package main

import (
	"context"

	"github.com/alexflint/go-arg"
)

// kwil-admin node ... (running node administration: ping, peer count, sql, )

type NodeCmd struct {
	RPCServer string `arg:"-s,--rpcserver" help:"RPC server address"`

	Ping *PingCmd `arg:"subcommand:ping" help:"Check connectivity with the nodes admin RPC interface"`
}

func (nc *NodeCmd) run(ctx context.Context) error {
	switch {
	case nc.Ping != nil:
		return nil // TODO -- need the authenticated gRPC service
	default:
		return arg.ErrHelp
	}
}

type PingCmd struct{}
