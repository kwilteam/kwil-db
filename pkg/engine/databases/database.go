package databases

import "github.com/kwilteam/kwil-db/pkg/sql/sqlite"

// A database is a single deployed instance of kwil-db.
// It contains a SQLite file
type Database struct {
	conn  *sqlite.Connection
	stmts map[string]*sqlite.Statement
}
