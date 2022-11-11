package server

import (
	"kwil/x/proto/apipb"
	"net/http"
	"os"
	"regexp"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcEndpointEnv = "KWIL_GRPC_ENDPOINT"
)

func Start() error {
	cmd := &cobra.Command{
		Use:   "api-gateway",
		Short: "api-gateway is a HTTP to gRPC gateway",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := runtime.NewServeMux()
			opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
			err := apipb.RegisterKwilServiceHandlerFromEndpoint(cmd.Context(), mux, viper.GetString("endpoint"), opts)
			if err != nil {
				return err
			}

			err = mux.HandlePath(http.MethodGet, "/api/v0/swagger.json", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				apipb.ServeSwaggerJSON(w, r)
			})
			if err != nil {
				return err
			}

			err = mux.HandlePath(http.MethodGet, "/swagger/ui", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
				apipb.ServeSwaggerUI(w, r)
			})

			if err != nil {
				return err
			}

			http_port := os.Getenv("GATEWAY_HTTP_PORT")
			if http_port == "" {
				http_port = ":8080"
			} else if http_port[0] != ':' {
				http_port = ":" + http_port
			}

			return http.ListenAndServe(http_port, cors(mux))
		},
	}

	grpc_url := os.Getenv("GRPC_CONTAINER_ENDPOINT")
	if grpc_url == "" {
		grpc_url = "localhost:50051"
	}

	cmd.PersistentFlags().String("endpoint", grpc_url, "gRPC server endpoint")
	err := viper.BindPFlag("endpoint", cmd.PersistentFlags().Lookup("endpoint"))
	if err != nil {
		return err
	}

	err = viper.BindEnv("endpoint", grpcEndpointEnv)
	if err != nil {
		return err
	}

	return cmd.Execute()
}

func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin(r.Header.Get("Origin")) {
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType")
		}

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func allowedOrigin(origin string) bool {
	cors := os.Getenv("GATEWAY_CORS")
	if cors == "" {
		cors = viper.GetString("cors")
	}
	if cors == "*" {
		return true
	}
	if matched, _ := regexp.MatchString(cors, origin); matched {
		return true
	}
	return false
}
