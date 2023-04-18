package server

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/internal/app/kwild/config"
	"kwil/internal/controller/http/v1/health"
	"kwil/internal/controller/http/v1/swagger"
	"kwil/internal/pkg/gateway/middleware"
	"kwil/pkg/log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GWServer struct {
	mux         *runtime.ServeMux
	middlewares []*middleware.NamedMiddleware
	logger      log.Logger
	h           http.Handler
	cfg         *config.KwildConfig
	httpServer  *http.Server
}

func NewGWServer(mux *runtime.ServeMux, cfg *config.KwildConfig, logger log.Logger) *GWServer {
	gw := &GWServer{mux: mux,
		logger: logger.Named("gateway"),
		h:      mux,
		cfg:    cfg,
	}

	gw.httpServer = &http.Server{
		Addr:    cfg.HttpListenAddress,
		Handler: gw,
	}
	return gw
}

func (g *GWServer) AddMiddlewares(ms ...*middleware.NamedMiddleware) {
	for _, m := range ms {
		g.middlewares = append(g.middlewares, m)
		g.logger.Info("apply middleware", zap.String("middleware", m.Name))
		g.h = m.Middleware(g.h)
	}
}

func (g *GWServer) Serve() error {
	g.logger.Info("kwil gateway started", zap.String("address", g.cfg.HttpListenAddress))
	return g.httpServer.ListenAndServe()
}

func (g *GWServer) Shutdown(ctx context.Context) error {
	return g.httpServer.Shutdown(ctx)
}

func (g *GWServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.h.ServeHTTP(w, r)
}

func (g *GWServer) SetupGrpcSvc(ctx context.Context) error {
	endpoint := g.cfg.GrpcListenAddress
	g.logger.Info("grpc address configured", zap.String("address", endpoint))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := txpb.RegisterTxServiceHandlerFromEndpoint(ctx, g.mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register tx service handler: %w", err)
	}

	return nil
}

func (g *GWServer) SetupHTTPSvc(ctx context.Context) error {
	err := g.mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", swagger.GWSwaggerJSONHandler)
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
