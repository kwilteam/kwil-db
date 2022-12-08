package hasura

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

func snakeCase(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "_", -1))
}

// customHasuraTableName return "schema_table".
func customHasuraTableName(schema, table string) string {
	names := []string{snakeCase(schema), snakeCase(table)}
	return strings.Join(names, "_")
}

// queryToExplain return a query body for explain API
// Does not support Directives yet.
func queryToExplain(query string) string {
	queryHead, queryBody, _ := strings.Cut(query, "{")
	queryHead = strings.Trim(queryHead, " ")
	s := strings.Split(queryHead, " ")
	if len(s) <= 1 {
		return fmt.Sprintf(`{"query": {"query": "{%s"}}`, queryBody)
	} else {
		operationName := s[1]
		return fmt.Sprintf(`{"query": {"query": "%s", "operationName": "%s"}}`, query, operationName)
	}
}

// InitializeHasura ensure Hasura is initialized, add default source('default') and schema
func InitializeHasura() {
	for {
		time.Sleep(3 * time.Second)
		client := NewClient(viper.GetString("graphql"))
		err := client.AddDefaultSourceAndSchema()
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			fmt.Println("wait for Hasura running...")
			continue
		}
		// ignore other error
		fmt.Println("Hasura initialized")
		break
	}
}
