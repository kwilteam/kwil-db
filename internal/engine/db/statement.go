package db

import "github.com/kwilteam/kwil-db/internal/sql"

type PreparedStatement struct {
	sql.Statement
	mutative bool
}

// IsMutative returns true if the statement changes the database.
// If false, the statement is read-only.
func (p *PreparedStatement) IsMutative() bool {
	return p.mutative
}
