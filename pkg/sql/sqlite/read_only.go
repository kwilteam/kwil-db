package sqlite

import (
	"strings"

	"github.com/kwilteam/go-sqlite"
)

// newReadOnlyConnection creates a new readonly connection.
// the connection returned from this function should NOT be returned from the package,
// and should only be used internally.
// it leaves out a lot of fields that are not needed for readonly connections.
func (c *Connection) newReadOnlyStatement(readConn *sqlite.Conn, stmt *sqlite.Stmt) *Statement {
	s := &Statement{
		conn: &Connection{
			conn: readConn,
			mu:   &nilMutex{},
			log:  *c.log.Named("readonly-connection"),
			name: c.name,
		},
		stmt: stmt,
	}

	s.determineColumnNames()

	return s
}

type nilMutex struct{}

func (m *nilMutex) Lock() {
	// adding in these panics in case this package changes later.
	// at the time of writing, it is impossible to get here.
	panic("nil mutex locked. this means you are likely using a readonly connection in a way that is not supported.  please see sqlite/read_only.go")
}
func (m *nilMutex) Unlock() {
	panic("nil mutex locked. this means you are likely using a readonly connection in a way that is not supported.  please see sqlite/read_only.go")
}

type lockable interface {
	Lock()
	Unlock()
}

func trimPadding(s string) string {
	ss := strings.TrimSpace(s)
	return ss
}
