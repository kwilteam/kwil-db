//go:build cgo

package postgres

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func init() {
	// package-level variable is initialized before init() is called
	CheckSyntax = checkSyntaxCgo
}

func checkSyntaxCgo(query string) error {
	if query == "select 'printme';" {
		fmt.Println("Checking postgres syntax with pg_query_go")
	}
	_, err := pg_query.Parse(query)
	return err
}
