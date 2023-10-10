package execution

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Engine is the struct encapsulating the engine.
type Engine struct {
	mu sync.Mutex

	// availableExtensions is a map of extension names to initializers.
	// It is provided when the engine is created.
	availableExtensions map[string]Initializer

	// procedures is a map of procedure names to procedures.
	procedures map[string]*Procedure

	// db is an interface to the datastore.
	db Datastore

	// loadCommand is the load command to execute when the engine is loaded.
	// it will be executed whenever the engine is created.
	loadCommand []*InstructionExecution

	// cache is a cache of initialized extensions and prepared statements.
	cache *cache

	// evaluater can evaluate expressions.
	// this is sort've a hack to support: my_ext.my_method($my_var + 1)
	evaluater Evaluater
}

type cache struct {
	initializedExtensions map[string]InitializedExtension
	preparedStatements    map[string]PreparedStatement
}

func NewEngine(ctx context.Context, db Datastore, eval Evaluater, opts *EngineOpts) (*Engine, error) {
	if opts == nil {
		opts = &EngineOpts{}
	}
	opts.ensureDefaults()

	procs := make(map[string]*Procedure)
	for name, proc := range opts.Procedures {
		procs[strings.ToLower(name)] = proc
	}

	exts := make(map[string]Initializer)
	for name, ext := range opts.Extensions {
		exts[strings.ToLower(name)] = ext
	}

	e := &Engine{
		mu:                  sync.Mutex{},
		availableExtensions: exts,
		procedures:          procs,
		loadCommand:         opts.LoadCmd,
		db:                  db,
		cache:               newCache(),
		evaluater:           eval,
	}

	err := e.executeLoad(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute load command: %w", err)
	}

	return e, nil
}

type EngineOpts struct {
	Extensions map[string]Initializer
	Procedures map[string]*Procedure
	LoadCmd    []*InstructionExecution
}

func (e *EngineOpts) ensureDefaults() {
	if e.Extensions == nil {
		e.Extensions = make(map[string]Initializer)
	}

	if e.Procedures == nil {
		e.Procedures = make(map[string]*Procedure)
	}

	if e.LoadCmd == nil {
		e.LoadCmd = make([]*InstructionExecution, 0)
	}
}

func newCache() *cache {
	return &cache{
		initializedExtensions: make(map[string]InitializedExtension),
		preparedStatements:    make(map[string]PreparedStatement),
	}
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var errs []string

	for _, stmt := range e.cache.preparedStatements {
		err := stmt.Close()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d prepared statements: %s", len(errs), strings.Join(errs, ", "))
	}

	return e.evaluater.Close()
}
