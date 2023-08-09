package engine

import (
	"github.com/kwilteam/kwil-db/pkg/log"
)

const (
	masterDBName = "kwil_master"
)

type EngineOpt func(*Engine)

// WithPath specifies the file path to which all sqlite databases will be written.
func WithPath(path string) EngineOpt {
	return func(e *Engine) {
		e.path = path
	}
}

// WithLogger allows the caller to specify a custom logger for the engine.
func WithLogger(l log.Logger) EngineOpt {
	return func(e *Engine) {
		e.log = l
	}
}

// WithMasterDBName allows the caller to specify a custom name for the master database file.
func WithMasterDBName(name string) EngineOpt {
	return func(e *Engine) {
		e.name = name
	}
}

// WithExtensions providers a map of extension initializers to the engine.
// Calling these initializers will return a new instance of the extension.
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

// WithOpener allows the caller to specify a custom sqlite opener for the engine.
// This is mostly useful for testing, where we want to teardown the database
func WithOpener(opener Opener) EngineOpt {
	return func(e *Engine) {
		e.opener = opener
	}
}

// ExecutionOpts are used to configure database execution.
type ExecutionOpts func(*executionConfig)

type executionConfig struct {
	// Sender is the address of the action caller.
	Sender string

	// ReadOnly is a flag that indicates if the execution is read-only.
	ReadOnly bool
}

// WithCaller sets the caller of the execution.
func WithCaller(caller string) ExecutionOpts {
	return func(cfg *executionConfig) {
		cfg.Sender = caller
	}
}

// ReadOnly sets the execution to read-only.
func ReadOnly(isReadOnly bool) ExecutionOpts {
	return func(cfg *executionConfig) {
		cfg.ReadOnly = isReadOnly
	}
}
