package graphql

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"kwil/x/graphql/hasura"
)

func CliSetup(cmd *cobra.Command) {
	cmd.PersistentFlags().String("graphql", "http://localhost:8082", "GraphQl server endpoint")
	viper.BindPFlag("graphql", cmd.PersistentFlags().Lookup("graphql"))
	viper.BindEnv("graphql", hasura.EndpointEnv)
	viper.BindEnv(hasura.AdminSecretName, hasura.AdminSecretEnv)
}
