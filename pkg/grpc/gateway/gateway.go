package gateway

import (
	"context"
	"net/http"

	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kwilteam/kwil-db/internal/controller/http/swagger"
	"github.com/kwilteam/kwil-db/internal/controller/http/v1/health"
	"github.com/kwilteam/kwil-db/pkg/grpc/gateway/middleware"
)

// GatewayServer is an HTTP server that can forward requests to GRPC servers.
type GatewayServer struct {
	http.Server
	mux         *runtime.ServeMux
	middlewares []*middleware.NamedMiddleware
	logger      log.Logger

	grpcServices []registeredGrpcService

	httpAddress string
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
		mux:         mux,
		logger:      log.NewNoOp(),
		httpAddress: httpAddress,

		grpcServices: []registeredGrpcService{},
	}

	for _, opt := range opts {
		opt(gw)
	}

	grpcDialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	for _, grpcService := range gw.grpcServices {
		if err := grpcService.connector(ctx, mux, grpcService.endpoint, grpcDialOpts); err != nil {
			return nil, err
		}
	}

	return gw, nil
}

// Start starts the gateway server
// This simply calls the HttpServer's ListenAndServe method, but is renamed for conformance with other servers
func (g *GatewayServer) Start() error {
	g.logger.Info("kwil gateway started", zap.String("address", g.httpAddress))
	return g.ListenAndServe()
}

func (g *GatewayServer) Shutdown(ctx context.Context) error {
	g.logger.Info("grpc gateway shutting down...")
	return g.Server.Shutdown(ctx)
}

func (g *GatewayServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *GatewayServer) SetupHTTPSvc(ctx context.Context) error {
	err := g.mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", swagger.GWSwaggerJSONV0Handler)
	if err != nil {
		return err
	}

	err = g.mux.HandlePath(http.MethodGet, "/api/v1/swagger.json", swagger.GWSwaggerJSONV1Handler)
	if err != nil {
		return err
	}

	err = g.mux.HandlePath(http.MethodGet, "/swagger/ui", swagger.GWSwaggerUIHandler)
	if err != nil {
		return err
	}

	// @yaiba TODO: https://grpc-ecosystem.github.io/grpc-gateway/docs/operations/health_check/
	err = g.mux.HandlePath(http.MethodGet, "/readyz", health.GWReadyzHandler)
	if err != nil {
		return err
	}
	err = g.mux.HandlePath(http.MethodGet, "/healthz", health.GWHealthzHandler)
	return err
}
