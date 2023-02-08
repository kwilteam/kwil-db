package server

import (
	"context"
	"fmt"
	accountpb "kwil/api/protobuf/kwil/account/v0/gen/go"
	pricingpb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/internal/app/kgw/config"
	"kwil/internal/controller/http/v0/health"
	"kwil/internal/pkg/gateway/middleware"
	"kwil/internal/pkg/graphql"
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
	cfg         config.AppConfig
}

func NewGWServer(mux *runtime.ServeMux, cfg config.AppConfig, logger log.Logger) *GWServer {
	return &GWServer{mux: mux,
		logger: logger,
		h:      mux,
		cfg:    cfg,
	}
}

func (g *GWServer) AddMiddlewares(ms ...*middleware.NamedMiddleware) {
	for _, m := range ms {
		g.middlewares = append(g.middlewares, m)
		g.logger.Info("apply middleware", zap.String("middleware", m.Name))
		g.h = m.Middleware(g.h)
	}
}

func (g *GWServer) Serve() error {
	g.logger.Info("kwil gateway started", zap.String("address", g.cfg.Server.Addr))
	return http.ListenAndServe(g.cfg.Server.Addr, g)
}

func (g *GWServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.h.ServeHTTP(w, r)
}

func (g *GWServer) SetupGrpcSvc(ctx context.Context) error {
	endpoint := g.cfg.Kwild.Endpoint
	g.logger.Info("grpc endpoint configured", zap.String("endpoint", endpoint))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := txpb.RegisterTxServiceHandlerFromEndpoint(ctx, g.mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register tx service handler: %w", err)
	}
	err = accountpb.RegisterAccountServiceHandlerFromEndpoint(ctx, g.mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register config service handler: %w", err)
	}
	err = pricingpb.RegisterPricingServiceHandlerFromEndpoint(ctx, g.mux, endpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register pricing service handler: %w", err)
	}

	return nil
}

func (g *GWServer) SetupHttpSvc(ctx context.Context) error {
	// @yaiba TODO: implement swagger api
	//err := g.mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", swagger.GWSwaggerJSONHandler)
	//if err != nil {
	//	return err
	//}
	//
	//err = g.mux.HandlePath(http.MethodGet, "/swagger/ui", swagger.GWSwaggerUIHandler)
	//
	//if err != nil {
	//	return err
	//}

	graphqlRProxy := graphql.NewRProxy(g.cfg.Graphql, g.logger.Named("rproxy"))
	err := g.mux.HandlePath(http.MethodPost, "/graphql", graphqlRProxy.Handler)
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
