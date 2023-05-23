package engine

import "github.com/kwilteam/kwil-db/pkg/log"

type MasterOpt func(*Engine)

func WithLogger(l log.Logger) MasterOpt {
	return func(m *Engine) {
		m.log = l
	}
}

func WithPath(path string) MasterOpt {
	return func(m *Engine) {
		m.path = path
	}
}
