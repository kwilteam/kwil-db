package client

import (
	"fmt"
	"os"

	"github.com/kwilteam/kwil-db/core/log"
)

var defaultPath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

type options struct {
	log  log.Logger
	path string
	name string
}

type SqliteOpts func(*options)

func WithPath(path string) SqliteOpts {
	return func(e *options) {
		e.path = path
	}
}

func WithLogger(l log.Logger) SqliteOpts {
	return func(e *options) {
		e.log = l
	}
}
