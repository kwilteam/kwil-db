package main

import (
	"context"
	"fmt"
	"kwil/x"
	"kwil/x/api/service"
	"kwil/x/cfgx"
	"kwil/x/messaging/mx"
	"kwil/x/messaging/pub"
	"net/http"
	"os"
	"regexp"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v0 "kwil/x/api/v0"
)

const (
	grpcEndpointEnv = "KWIL_GRPC_ENDPOINT"
)

func run() error {
	ctx, err := setupRootRequestCtx()
	if err != nil {
		return err
	}

	cmd := &cobra.Command{
		Use:   "api-gateway",
		Short: "api-gateway is a gRPC to HTTP gateway",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := runtime.NewServeMux()
			opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
			err := v0.RegisterKwilServiceHandlerFromEndpoint(cmd.Context(), mux, viper.GetString("endpoint"), opts)
			if err != nil {
				return err
			}

			err = mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				v0.ServeSwaggerJSON(w, r)
			})
			if err != nil {
				return err
			}

			err = mux.HandlePath(http.MethodGet, "/swagger/ui", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				v0.ServeSwaggerUI(w, r)
			})
			if err != nil {
				return err
			}

			return http.ListenAndServe(":8080", cors(ctx, mux))
		},
	}

	cmd.PersistentFlags().String("endpoint", "localhost:50051", "gRPC server endpoint")
	err = viper.BindPFlag("endpoint", cmd.PersistentFlags().Lookup("endpoint"))
	if err != nil {
		return err
	}

	err = viper.BindEnv("endpoint", grpcEndpointEnv)
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func cors(ctx context.Context, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin(r.Header.Get("Origin")) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
		}

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func allowedOrigin(origin string) bool {
	if viper.GetString("cors") == "*" {
		return true
	}
	if matched, _ := regexp.MatchString(viper.GetString("cors"), origin); matched {
		return true
	}
	return false
}

func setupRootRequestCtx() (context.Context, error) {
	cfg := cfgx.GetTestConfig().Select("messaging-emitter")

	// Once the message type is known, we will create the
	// appropriate serdes
	serdes := mx.SerdesByteArray()

	// Using NewEmitterSingleClient for now. Once we need
	// more than one emitter, we will need to create the client
	// separately and close it our upon application shutdown.
	e, err := pub.NewEmitterSingleClient(cfg, serdes)
	if err != nil {
		return nil, err
	}

	// Not sure where else to inject a service for down stream consumption
	return x.Wrap(x.RootContext(), service.DATABASE_EMITTER_ALIAS, e), nil
}
