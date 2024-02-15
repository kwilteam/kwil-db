//go:build enablecgo

package postgres

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func init() {
	// package-level variable is initialized before init() is called
	CheckSyntax = checkSyntaxCgo
}

func checkSyntaxCgo(query string) error {
	_, err := pg_query.Parse(query)
	return err
}
