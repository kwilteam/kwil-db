package server

import (
	kg "kwil/cmd/kwil-gateway/server"
	"kwil/cmd/kwild/common"
	"kwil/pkg/chain/types"
	"kwil/pkg/logger"
	"kwil/x/async"
	"os"

	"github.com/spf13/cobra"
)

func Start() error {
	cmd := &cobra.Command{
		Use:   "kwild",
		Short: "kwil grpc server",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			logger := logger.New()

			stop := func(err error) {
				logger.Sugar().Error(err)
				os.Exit(1)
			}

			kwild := func() error {
				return execute(logger)
			}

			if !isGatewayEnabled() {
				if err := kwild(); err != nil {
					stop(err)
				}
			}

			async.Run(kg.Start).Catch(stop)

			<-async.Run(kwild).Catch(stop).DoneCh()
		},
	}

	common.BindKwildFlags(cmd)
	common.BindKwildEnv(cmd)
	types.BindChainFlags(cmd.PersistentFlags())
	types.BindChainEnv()

	return cmd.Execute()
}
