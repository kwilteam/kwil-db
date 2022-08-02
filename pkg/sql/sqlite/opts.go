package sqlite

import (
	"fmt"
	"os"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/log"
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

// WithAttachedDatabase
func WithAttachedDatabase(name string, fileName string) ConnectionOption {
	return func(conn *Connection) {
		name = strings.ToLower(name)

		if _, ok := conn.attachedDBs[name]; ok {
			panic("attached database already exists: " + name)
		}

		conn.attachedDBs[name] = fileName
	}
}
