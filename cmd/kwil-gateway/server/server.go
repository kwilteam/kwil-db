package server

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gateway2 "kwil/internal/pkg/gateway"
	auth2 "kwil/internal/pkg/gateway/middleware/auth"
	"kwil/internal/pkg/gateway/middleware/cors"
	"os"
)

func Start() error {
	cmd := &cobra.Command{
		Use:   "kwil-gateway",
		Short: "gateway to kwil service",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			mux := runtime.NewServeMux()
			addr := viper.GetString(gateway2.ListenAddressFlag)
			gw := gateway2.NewGWServer(mux, addr)

			if err := gw.SetupGrpcSvc(cmd.Context()); err != nil {
				return err
			}
			if err := gw.SetupHttpSvc(cmd.Context()); err != nil {
				return err
			}

			f, err := os.Open("keys.json")
			if err != nil {
				return err
			}

			keyManager, err := auth2.NewKeyManager(f)
			if err != nil {
				return err
			}
			f.Close()

			_cors := viper.GetString(cors.GatewayCorsFlag)
			gw.AddMiddlewares(
				// from innermost middleware
				auth2.MAuth(keyManager),
				cors.MCors(_cors),
			)

			return gw.Serve()
		},
	}

	BindGatewayFlags(cmd)
	return cmd.Execute()
}
