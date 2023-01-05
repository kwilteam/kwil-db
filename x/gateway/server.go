package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kwil/x/gateway/middleware"
	"kwil/x/graphql"
	"kwil/x/logx"
	"kwil/x/proto/apipb"
	"net/http"
)

type GWServer struct {
	mux         *runtime.ServeMux
	addr        string
	middlewares []*middleware.NamedMiddleware
	logger      logx.Logger
	h           http.Handler
}

func NewGWServer(mux *runtime.ServeMux, addr string) *GWServer {
	logger := logx.New()
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
	g.logger.Info("grpc endpoint configured", zap.String("endpoint", viper.GetString(GrpcEndpointName)))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	return apipb.RegisterKwilServiceHandlerFromEndpoint(ctx, g.mux, viper.GetString(GrpcEndpointName), opts)
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
