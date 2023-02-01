package server

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"kwil/internal/pkg/gateway"
	"kwil/internal/pkg/gateway/middleware/auth"
	"kwil/internal/pkg/gateway/middleware/cors"
	"kwil/internal/pkg/graphql/hasura"
)

const (
	defaultGrpcEndpoint    = "localhost:50051"
	defaultGraphqlEndpoint = "http://localhost:8080"
	defaultAddr            = "0.0.0.0:8082"
	defaultCors            = ""
)

func BindGatewayEnv() {
	viper.BindEnv(gateway.GrpcEndpointFlag, gateway.GrpcEndpointEnv)
	viper.BindEnv(gateway.ListenAddressFlag, gateway.ListenAddressEnv)

	viper.BindEnv(hasura.GraphqlEndpointFlag, hasura.GraphqlEndpointEnv)
	viper.BindEnv(hasura.AdminSecretFlag, hasura.AdminSecretEnv)

	viper.BindEnv(cors.GatewayCorsFlag, cors.GatewayCorsEnv)
	viper.BindEnv(auth.HealthCheckApiKeyValueFlag, auth.HealthCheckApiKeyValueEnv)
}

func BindGatewayFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(gateway.GrpcEndpointFlag, defaultGrpcEndpoint, "kwild gRPC server endpoint")
	viper.BindPFlag(gateway.GrpcEndpointFlag, cmd.PersistentFlags().Lookup(gateway.GrpcEndpointFlag))

	cmd.PersistentFlags().String(gateway.ListenAddressFlag, defaultAddr, "gateway server listen address")
	viper.BindPFlag(gateway.ListenAddressFlag, cmd.PersistentFlags().Lookup(gateway.ListenAddressFlag))

	cmd.PersistentFlags().String(hasura.GraphqlEndpointFlag, defaultGraphqlEndpoint, "GraphQl server endpoint")
	viper.BindPFlag(hasura.GraphqlEndpointFlag, cmd.PersistentFlags().Lookup(hasura.GraphqlEndpointFlag))

	cmd.PersistentFlags().String(cors.GatewayCorsFlag, defaultCors, "gateway CORS setting, list separated by commas")
	viper.BindPFlag(cors.GatewayCorsFlag, cmd.PersistentFlags().Lookup(cors.GatewayCorsFlag))

	BindGatewayEnv()
}
