package engine

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/extensions"
	"github.com/kwilteam/kwil-db/pkg/log"
)

const (
	defaultName = "kwil_master"
)

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

func WithFileName(name string) EngineOpt {
	return func(e *engine) {
		e.name = name
	}
}

func WithWipe() EngineOpt {
	return func(e *engine) {
		e.wipeOnStart = true
	}
}

func WithExtension(name, endpoint string, config map[string]string) EngineOpt {
	return func(e *engine) {
		lowerName := strings.ToLower(name)
		if _, ok := e.extensions[lowerName]; ok {
			panic("extension of same name already registered: " + name)
		}

		e.extensions[lowerName] = extensions.New(name, endpoint, config)
	}
}
