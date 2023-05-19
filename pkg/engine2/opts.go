package engine2

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

type EngineOpt func(*engine)

func WithPath(path string) EngineOpt {
	return func(e *engine) {
		e.path = path
	}
}

func WithLogger(l log.Logger) EngineOpt {
	return func(e *engine) {
		e.log = l
	}
}
