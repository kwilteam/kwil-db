package engine

import (
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

func WithName(name string) EngineOpt {
	return func(e *engine) {
		e.name = name
	}
}

func WithWipe() EngineOpt {
	return func(e *engine) {
		e.wipeOnStart = true
	}
}
