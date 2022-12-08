package server

import (
	"encoding/json"
	"io"
	"kwil/x/graphql"
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

			graphqlRProxy := graphql.NewGraphqlRProxy()
			err = mux.HandlePath(http.MethodPost, "/graphql", graphqlRProxy.Handler)
			if err != nil {
				return err
			}

			http_port := os.Getenv("GATEWAY_HTTP_PORT")
			if http_port == "" {
				http_port = ":8080"
			} else if http_port[0] != ':' {
				http_port = ":" + http_port
			}

			// add the api key middleware
			apik, err := newApiKeyMiddleware("keys.json")
			if err != nil {
				return err
			}

			return http.ListenAndServe(http_port, apik(cors(mux)))
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

	graphql.CliSetup(cmd)

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

// since I am making this on the last day, a lot of this is hacked together and probably not super optimal
type keyJson struct {
	Keys []string `json:"keys"`
}

// reads in the path and loads the keys
func newApiKeyMiddleware(path string) (func(http.Handler) http.Handler, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bts, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var keys keyJson
	err = json.Unmarshal(bts, &keys)
	if err != nil {
		return nil, err
	}

	km := make(map[string]struct{})
	for _, k := range keys.Keys {
		km[k] = struct{}{}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := km[r.Header.Get("x-api-key")]; ok {
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		})
	}, nil
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
