package gateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/services/grpc_gateway/middleware"
	"google.golang.org/grpc"
)

type GatewayOpt func(*GatewayServer)

// WithMiddleware adds a middleware to the gateway server.
func WithMiddleware(m *middleware.NamedMiddleware) GatewayOpt {
	return func(gw *GatewayServer) {
		gw.middlewares = append(gw.middlewares, m)
	}
}

// WithLogger sets the logger for the gateway server.
func WithLogger(logger log.Logger) GatewayOpt {
	return func(gw *GatewayServer) {
		gw.logger = logger
	}
}

// GrpcConnector is a function that connects the gateway to a GRPC server.
type GrpcConnector func(ctx context.Context, mux *runtime.ServeMux, address string, opts []grpc.DialOption) error

// WithGrpcService registers a grpc service with the gateway server.
func WithGrpcService(endpoint string, connector GrpcConnector) GatewayOpt {
	return func(gw *GatewayServer) {
		gw.grpcServices = append(gw.grpcServices, registeredGrpcService{
			endpoint:  endpoint,
			connector: connector,
		})
	}
}
