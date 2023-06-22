package sqldb

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/pkg/log"
)

var defaultPath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

type SqliteOpts func(*SqliteStore)

func WithPath(path string) SqliteOpts {
	return func(e *SqliteStore) {
		e.path = path
	}
}

func WithLogger(l log.Logger) SqliteOpts {
	return func(e *SqliteStore) {
		e.log = l
	}
}

func WithGlobalVariables(vars map[string]any) SqliteOpts {
	return func(e *SqliteStore) {
		e.globalVars = vars
	}
}
