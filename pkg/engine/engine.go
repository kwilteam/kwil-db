package engine

import (
	"fmt"
	"kwil/pkg/engine/datasets"
	"kwil/pkg/log"
	"kwil/pkg/sql/driver"
	"os"
	"strings"
)

// the master is used to connect to the master sqlite database.
// it can find other databases, and it can create new databases.
type Engine struct {
	conn     *driver.Connection
	path     string
	log      log.Logger
	Datasets map[string]*datasets.Dataset
}

// Open opens the master database and loads all datasets into memory.
func Open(opts ...MasterOpt) (*Engine, error) {
	e := &Engine{
		log:      log.NewNoOp(),
		path:     DefaultPath,
		Datasets: make(map[string]*datasets.Dataset),
	}
	for _, opt := range opts {
		opt(e)
	}

	var err error
	e.conn, err = driver.OpenConn("master", driver.WithPath(e.path))
	if err != nil {
		return nil, fmt.Errorf("failed to open master connection: %w", err)
	}

	err = e.conn.AcquireLock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	err = e.initTables()
	if err != nil {
		return nil, fmt.Errorf("failed to init: %w", err)
	}

	err = e.validateRegisteredDatasets()
	if err != nil {
		return nil, fmt.Errorf("failed to validate registered datasets: %w", err)
	}

	err = e.loadAllDataSets()
	if err != nil {
		return nil, fmt.Errorf("failed to load all datasets: %w", err)
	}

	return e, nil
}

// Close closes the master database connection.
func (e *Engine) Close() error {
	return e.conn.Close()
}

// CreateDataset creates a new dataset in the master database, on disk, and in memory.
func (e *Engine) CreateDataset(owner, name string) error {
	if _, ok := e.Datasets[name]; ok {
		return fmt.Errorf("dataset already exists")
	}

	err := e.addDatasetToMaster(owner, name)
	if err != nil {
		return fmt.Errorf("failed to add dataset to master: %w", err)
	}

	dataset, err := datasets.OpenDataset(owner, name, e.path)
	if err != nil {
		return fmt.Errorf("failed to open dataset: %w", err)
	}

	e.Datasets[dataset.DBID] = dataset

	return nil
}

// DeleteDataset deletes a dataset from the master database, disk, and memory.
// it starts with memory, then disk, then master.
// the order is important because if it fails between disk and master,
// it will catch it when it runs "validateRegisteredDatasets"
func (e *Engine) DeleteDataset(dbid string) error {
	delete(e.Datasets, dbid)

	err := e.deleteDatasetFromDisk(dbid)
	if err != nil {
		return fmt.Errorf("failed to delete dataset from disk: %w", err)
	}

	err = e.deleteDatasetFromMaster(dbid)
	if err != nil {
		return fmt.Errorf("failed to delete dataset from master: %w", err)
	}

	return nil
}

// initTables initializes the master database tables.
func (e *Engine) initTables() error {
	return e.conn.Execute(sqlInitTables)
}

// loadAllDataSets loads all datasets from the master database into memory.
func (e *Engine) loadAllDataSets() error {
	err := e.conn.Query(sqlListDatabases, func(stmt *driver.Statement) error {
		owner := stmt.GetText("owner")
		name := stmt.GetText("name")

		dataset, err := datasets.OpenDataset(name, owner, e.path)
		if err != nil {
			return fmt.Errorf("failed to open dataset of name: %s, owner: %s: %w", name, owner, err)
		}

		e.Datasets[dataset.DBID] = dataset

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// validateRegisteredDatasets checks that all datasets in the master database
// exist on disk. If they don't, they are removed from the master database.
func (e *Engine) validateRegisteredDatasets() error {
	dbids := make([]string, 0)
	err := e.conn.Query(sqlListDatabases, func(stmt *driver.Statement) error {
		dbids = append(dbids, stmt.GetText("dbid"))
		return nil
	})
	if err != nil {
		return err
	}

	files, err := readDir(e.path)
	if err != nil {
		return err
	}

	fileNames := make(map[string]struct{})
	for _, file := range files {
		fileNames[strings.TrimSuffix(file.Name(), fileSuffix)] = struct{}{}
	}

	for _, dbid := range dbids {
		if _, ok := fileNames[dbid]; !ok {
			err = e.deleteDatasetFromMaster(dbid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func readDir(dirPath string) ([]os.FileInfo, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return files, nil
}
