package misc

import (
	"kwil/x/graphql/hasura"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultGraphqlEndpoint = "http://localhost:8082"
)

func CliSetup(cmd *cobra.Command) {
	cmd.PersistentFlags().String(hasura.GraphqlEndpointName, defaultGraphqlEndpoint, "GraphQl server endpoint")
	viper.BindPFlag(hasura.GraphqlEndpointName, cmd.PersistentFlags().Lookup("graphql"))
	CliBindEnv()
}

func CliBindEnv() {
	viper.BindEnv(hasura.GraphqlEndpointName, hasura.EndpointEnv)
	viper.BindEnv(hasura.AdminSecretName, hasura.AdminSecretEnv)
}
