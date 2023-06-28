package engine

import (
	"github.com/kwilteam/kwil-db/pkg/log"
)

const (
	masterDBName = "kwil_master"
)

type EngineOpt func(*Engine)

func WithPath(path string) EngineOpt {
	return func(e *Engine) {
		e.path = path
	}
}

func WithLogger(l log.Logger) EngineOpt {
	return func(e *Engine) {
		e.log = l
	}
}

func WithMasterDBName(name string) EngineOpt {
	return func(e *Engine) {
		e.name = name
	}
}
func WithExtensions(exts map[string]ExtensionInitializer) EngineOpt {
	return func(e *Engine) {
		for name, ext := range exts {

			if _, ok := e.extensions[name]; ok {
				panic("extension of same name already registered: " + name)
			}

			e.extensions[name] = ext
		}
	}
}
