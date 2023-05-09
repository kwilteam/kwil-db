package sqlite

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/log"
	"os"

	"github.com/kwilteam/go-sqlite"
)

var (
	DefaultPath string
)

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "./tmp"
	}

	DefaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

type ConnectionOption func(*Connection)

// WithLogger specifies the logger to use
func WithLogger(logger log.Logger) ConnectionOption {
	return func(conn *Connection) {
		conn.log = logger
	}
}

// WithPath specifies the path to the sqlite database
func WithPath(path string) ConnectionOption {
	return func(conn *Connection) {
		conn.path = path
	}
}

// WithConnectionPoolSize specifies the size of the pool of readonly connections.
// We restrict ReadWrite connections to only having 1 to help prevent non-determinism between systems
func WithConnectionPoolSize(size int) ConnectionOption {
	return func(c *Connection) {
		c.poolSize = size
	}
}

// WithGlobalVariables adds global variables to the connection
func WithGlobalVariables(globalVariables map[string]any) ConnectionOption {
	return func(conn *Connection) {
		for name, value := range globalVariables {
			if _, ok := conn.globalVariableMap[name]; ok {
				panic("global variable already exists: " + name)
			}

			conn.globalVariableMap[name] = value
		}
	}
}

// InMemory specifies that the database should be in memory.
// WARNING: If the database in in-memory, then users will be able to execute ad-hoc queries that would
// otherwise be read-only
func InMemory() ConnectionOption {
	return func(conn *Connection) {
		conn.path = "file::memory:?mode=memory"
		// need to disable wal mode and use shared cache for in memory databases
		// also need to enable URI mode since the in-memory database is not a file
		conn.flags = conn.flags | sqlite.OpenSharedCache | sqlite.OpenURI&^sqlite.OpenWAL
		conn.isMemory = true
	}
}
