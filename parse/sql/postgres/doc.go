/*
Package postgres provides a CheckSyntax function wraps the parser for the
PostgreSQL dialect of SQL, which will return an error if the query is not
syntactically valid.

This package includes a default implementation of CheckSyntax that does nothing,
and a cgo implementation that uses the pg_query_go library to check the syntax
of the query.

This cgo implementation is only built when the `enablecgo` build tag is set.
By doing this, we keep pg_query_go as a cgo dependency, which is not required in
the default build.

pg_query_go is a cgo wrapper around the libpg_query C library, which is used in
kwil-db to simply check the syntax of SQL queries we generated.

TO run tests using pg_query_go, run `go test -tags enablecgo ./...`
*/
package postgres
