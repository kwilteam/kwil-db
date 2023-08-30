package engine

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/log"
)

const (
	masterDBName = "kwil_master"
)

type EngineOpt func(*Engine)

// WithLogger allows the caller to specify a custom logger for the engine.
func WithLogger(l log.Logger) EngineOpt {
	return func(e *Engine) {
		e.log = l
	}
}

// WithExtensions providers a map of extension initializers to the engine.
// Calling these initializers will return a new instance of the extension.
func WithExtensions(exts map[string]ExtensionInitializer) EngineOpt {
	return func(e *Engine) {
		for name, ext := range exts {
			lowerName := strings.ToLower(name)
			if _, ok := e.extensions[lowerName]; ok {
				panic("extension of same name already registered: " + lowerName)
			}

			e.extensions[lowerName] = ext
		}
	}
}

// ExecutionOpt are used to configure database execution.
type ExecutionOpt func(*executionConfig)

type executionConfig struct {
	// Sender is the address of the action caller.
	Sender types.UserIdentifier

	// ReadOnly is a flag that indicates if the execution is read-only.
	ReadOnly bool
}

// WithCaller sets the caller of the execution.
func WithCaller(caller types.UserIdentifier) ExecutionOpt {
	return func(cfg *executionConfig) {
		cfg.Sender = caller
	}
}

// ReadOnly sets the execution to read-only.
func ReadOnly(isReadOnly bool) ExecutionOpt {
	return func(cfg *executionConfig) {
		cfg.ReadOnly = isReadOnly
	}
}
