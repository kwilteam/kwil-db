package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DefaultGRPCOpts returns the default grpc options for the client.
func DefaultGRPCOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(64 * 1024 * 1024), // 64MiB limit on *responses*; sends are unlimited
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}
