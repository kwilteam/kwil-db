package server

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"kwil/x/gateway"
	"kwil/x/gateway/middleware/cors"
	"kwil/x/graphql/hasura"
)

const (
	defaultGrpcEndpoint    = "localhost:50051"
	defaultGraphqlEndpoint = "http://localhost:8080"
	defaultAddr            = "0.0.0.0:8082"
	defaultCors            = ""
)

func CliBindEnv() {
	viper.BindEnv(gateway.GrpcEndpointName, gateway.GrpcEndpointEnv)
	viper.BindEnv(gateway.ListenAddressName, gateway.ListenAddressEnv)

	viper.BindEnv(hasura.GraphqlEndpointName, hasura.EndpointEnv)
	viper.BindEnv(hasura.AdminSecretName, hasura.AdminSecretEnv)

	viper.BindEnv(cors.GatewayCorsName, cors.GatewayCorsEnv)
}

func CliSetup(cmd *cobra.Command) {
	cmd.PersistentFlags().String(gateway.GrpcEndpointName, defaultGrpcEndpoint, "kwild gRPC server endpoint")
	viper.BindPFlag(gateway.GrpcEndpointName, cmd.PersistentFlags().Lookup(gateway.GrpcEndpointName))

	cmd.PersistentFlags().String(gateway.ListenAddressName, defaultAddr, "gateway server listen address")
	viper.BindPFlag(gateway.ListenAddressName, cmd.PersistentFlags().Lookup(gateway.ListenAddressName))

	cmd.PersistentFlags().String(hasura.GraphqlEndpointName, defaultGraphqlEndpoint, "GraphQl server endpoint")
	viper.BindPFlag(hasura.GraphqlEndpointName, cmd.PersistentFlags().Lookup(hasura.GraphqlEndpointName))

	cmd.PersistentFlags().String(cors.GatewayCorsName, defaultCors, "gateway CORS setting, list separated by commas")
	viper.BindPFlag(cors.GatewayCorsName, cmd.PersistentFlags().Lookup(cors.GatewayCorsName))

	CliBindEnv()
}
