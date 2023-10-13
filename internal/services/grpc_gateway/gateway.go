package gateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/services/grpc_gateway/middleware"
	"github.com/kwilteam/kwil-db/internal/services/http/health"
	"github.com/kwilteam/kwil-db/internal/services/http/swagger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GatewayServer is an HTTP server that can forward requests to GRPC servers.
type GatewayServer struct {
	http.Server
	middlewares  []*middleware.NamedMiddleware
	grpcServices []registeredGrpcService
	logger       log.Logger
}

type registeredGrpcService struct {
	endpoint  string
	connector GrpcConnector
}

func NewGateway(ctx context.Context, httpAddress string, opts ...GatewayOpt) (*GatewayServer, error) {
	// I moved this internally since this gateway is pretty coupled to the GRPC Gateway v2
	// the main point of the gateway is to forward to the GRPC server
	mux := runtime.NewServeMux()

	gw := &GatewayServer{
		Server: http.Server{
			Addr:    httpAddress,
			Handler: mux,
		},
		logger:       log.NewNoOp(),
		grpcServices: []registeredGrpcService{},
	}

	for _, opt := range opts {
		opt(gw)
	}

	for _, m := range gw.middlewares {
		gw.logger.Info("apply middleware", zap.String("name", m.Name))
		gw.Handler = m.Middleware(gw.Handler)
	}

	grpcDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	for _, grpcService := range gw.grpcServices {
		if err := grpcService.connector(ctx, mux, grpcService.endpoint, grpcDialOpts); err != nil {
			return nil, err
		}
	}

	gw.logger.Info("register extra helper endpoints")
	err := registerHelperEndpoints(mux)
	if err != nil {
		return nil, fmt.Errorf("error register extra endpoints: %w", err)
	}

	return gw, nil
}

// Start starts the gateway server
// This simply calls the HttpServer's ListenAndServe method, but is renamed for conformance with other servers
func (g *GatewayServer) Start() error {
	g.logger.Info("gateway server started", zap.String("address", g.Server.Addr))
	return g.ListenAndServe()
}

func (g *GatewayServer) Shutdown(ctx context.Context) error {
	g.logger.Info("gateway server shutting down...")
	return g.Server.Shutdown(ctx)
}

func registerHelperEndpoints(mux *runtime.ServeMux) error {
	// err := mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", swagger.GWSwaggerJSONV0Handler)
	// if err != nil {
	// 	return err
	// }

	err := mux.HandlePath(http.MethodGet, "/api/v1/swagger.json", swagger.GWSwaggerJSONV1Handler)
	if err != nil {
		return err
	}

	err = mux.HandlePath(http.MethodGet, "/swagger/ui", swagger.GWSwaggerUIHandler)
	if err != nil {
		return err
	}

	// @yaiba TODO: https://grpc-ecosystem.github.io/grpc-gateway/docs/operations/health_check/
	err = mux.HandlePath(http.MethodGet, "/readyz", health.GWReadyzHandler)
	if err != nil {
		return err
	}
	err = mux.HandlePath(http.MethodGet, "/healthz", health.GWHealthzHandler)
	return err
}
