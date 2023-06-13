package extensions

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/engine/sqldb/sqlite"
)

// ExtensionRunner runs extensions concurrently.
type ExtensionRunner struct {
	extensions map[string]*Extension
}

type extensionDatastoreOpener func(string) (*sqlite.SqliteStore, error)

// Initialize initializes all extensions.
// It will initialize each concurrently, but will wait for all to finish before returning.
func (e *ExtensionRunner) Initialize(ctx context.Context, opener extensionDatastoreOpener) error {
	errChan := make(chan error, len(e.extensions))
	var wg sync.WaitGroup

	for _, ext := range e.extensions {
		wg.Add(1)
		go e.runExtensionInitialization(ctx, opener, ext, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("error initializing extensions: %v", errors)
	}

	return nil
}

// runExtensionInitialization runs the initialization for the given extension.
func (e *ExtensionRunner) runExtensionInitialization(ctx context.Context, opener extensionDatastoreOpener, ext *Extension, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()
	var err error
	defer func() {
		// recover from panic from extension
		if r := recover(); r != nil {
			if err == nil {
				err = fmt.Errorf("panic running extension %s: %v", ext.Name, r)
			}
		}

		if err != nil {
			errChan <- err
		}
	}()

	ds, err := opener(ext.Name)
	if err != nil {
		err = fmt.Errorf("error opening datastore for extension %s: %w", ext.Name, err)
		return
	}

	savepoint, err := ds.Savepoint()
	if err != nil {
		err = fmt.Errorf("error creating savepoint for extension %s: %w", ext.Name, err)
		return
	}
	defer savepoint.Rollback()

	for _, table := range ext.Tables {
		if err = ds.CreateTable(ctx, table); err != nil {
			err = fmt.Errorf("error creating table %s for extension %s: %w", table.Name, ext.Name, err)
			return
		}
	}

	if err = ext.Initialize(ds); err != nil {
		err = fmt.Errorf("error initializing extension %s: %w", ext.Name, err)
		return
	}

	if err = savepoint.Commit(); err != nil {
		err = fmt.Errorf("error committing savepoint for extension %s: %w", ext.Name, err)
		return
	}
}

// RunAll runs all extension cron jobs concurrently.
func (e *ExtensionRunner) RunAll(bts map[string][]byte) error {
	errChan := make(chan error, len(e.extensions))

	wg := sync.WaitGroup{}
	for name, bts := range bts {
		wg.Add(1)
		name := name
		bts := bts

		go func() {
			defer wg.Done()

			err := e.Run(name, bts)
			if err != nil {
				errChan <- fmt.Errorf("error running extension %s: %w", name, err)
			}
		}()
	}

	wg.Wait()

	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("error running extensions: %v", errors)
	}

	return nil
}

// Run runs the cron job for the given extension.
func (e *ExtensionRunner) Run(name string, bts []byte) error {
	ext, ok := e.extensions[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("extension %s not found", name)
	}

	return ext.RunCron(bts)
}

// Register registers an extension.
func (e *ExtensionRunner) Register(ext *Extension) error {
	if e.extensions == nil {
		e.extensions = make(map[string]*Extension)
	}

	lowerName := strings.ToLower(ext.Name)

	if _, ok := e.extensions[lowerName]; ok {
		return fmt.Errorf("extension with name %s already registered", ext.Name)
	}

	if _, ok := reservedExtensionNames[lowerName]; ok {
		return fmt.Errorf("extension %s is using a reserved name", ext.Name)
	}

	e.extensions[lowerName] = ext
	return nil
}

// GetHeaders gets the headers for all extensions.
func (e *ExtensionRunner) GetHeaders() (map[string][]byte, error) {
	headers := make(map[string][]byte)

	for name, ext := range e.extensions {
		header, err := ext.GetHeader()
		if err != nil {
			return nil, fmt.Errorf("error getting header for extension %s: %w", name, err)
		}

		headers[name] = header
	}

	return headers, nil
}

var (
	reservedExtensionNames = map[string]struct{}{
		"kwil_master": {},
		"accounts_db": {},
	}
)
