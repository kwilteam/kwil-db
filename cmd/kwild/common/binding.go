package common

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"kwil/internal/pkg/graphql/hasura"
)

const (
	PgDatabaseUrlFlag = "pg-database-url"
	PgDatabaseUrlEnv  = "PG_DATABASE_URL"
)

func BindKwildFlags(cmd *cobra.Command) {
	fs := cmd.PersistentFlags()

	fs.String(hasura.GraphqlEndpointFlag, "", "the endpoint of the Hasura node")
	viper.BindPFlag(hasura.GraphqlEndpointFlag, fs.Lookup(hasura.GraphqlEndpointFlag))

}

func BindKwildEnv(cmd *cobra.Command) {
	viper.BindEnv(hasura.GraphqlEndpointFlag, hasura.GraphqlEndpointEnv)
}
