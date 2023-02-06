package gateway

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/api/protobuf/v0/pb/accountspb"
	"kwil/api/protobuf/v0/pb/pricingpb"
	"kwil/internal/pkg/gateway/middleware"
	"kwil/internal/pkg/graphql"
	"kwil/pkg/logger"
	"kwil/x/proto/apipb"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GWServer struct {
	mux         *runtime.ServeMux
	addr        string
	middlewares []*middleware.NamedMiddleware
	logger      logger.Logger
	h           http.Handler
}

func NewGWServer(mux *runtime.ServeMux, addr string) *GWServer {
	logger := logger.New()
	return &GWServer{mux: mux,
		addr:   addr,
		logger: logger,
		h:      mux,
	}
}

func (g *GWServer) AddMiddlewares(ms ...*middleware.NamedMiddleware) {
	for _, m := range ms {
		g.middlewares = append(g.middlewares, m)
		g.logger.Info("apply middleware", zap.String("name", m.Name))
		g.h = m.Middleware(g.h)
	}
}

func (g *GWServer) Serve() error {
	g.logger.Info("gateway started", zap.String("address", g.addr))
	return http.ListenAndServe(g.addr, g)
}

func (g *GWServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.h.ServeHTTP(w, r)
}

func (g *GWServer) SetupGrpcSvc(ctx context.Context) error {
	g.logger.Info("grpc endpoint configured", zap.String("endpoint", viper.GetString(GrpcEndpointFlag)))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := txpb.RegisterTxServiceHandlerFromEndpoint(ctx, g.mux, viper.GetString(GrpcEndpointFlag), opts)
	if err != nil {
		return fmt.Errorf("failed to register tx service handler: %w", err)
	}
	err = accountspb.RegisterAccountServiceHandlerFromEndpoint(ctx, g.mux, viper.GetString(GrpcEndpointFlag), opts)
	if err != nil {
		return fmt.Errorf("failed to register config service handler: %w", err)
	}
	err = pricingpb.RegisterPricingServiceHandlerFromEndpoint(ctx, g.mux, viper.GetString(GrpcEndpointFlag), opts)
	if err != nil {
		return fmt.Errorf("failed to register pricing service handler: %w", err)
	}

	return nil
}

func (g *GWServer) SetupHttpSvc(ctx context.Context) error {
	err := g.mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		apipb.ServeSwaggerJSON(w, r)
	})
	if err != nil {
		return err
	}

	err = g.mux.HandlePath(http.MethodGet, "/swagger/ui", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		apipb.ServeSwaggerUI(w, r)
	})

	if err != nil {
		return err
	}

	graphqlRProxy := graphql.NewRProxy()
	err = g.mux.HandlePath(http.MethodPost, "/graphql", graphqlRProxy.Handler)
	if err != nil {
		return err
	}

	err = g.mux.HandlePath(http.MethodGet, "/readyz", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// TODO: check dependency?
		w.WriteHeader(http.StatusOK)
	})

	err = g.mux.HandlePath(http.MethodGet, "/healthz", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		// won't check dependent services
		w.WriteHeader(http.StatusOK)
	})

	return err
}
