package engine

import (
	"fmt"
	"kwil/pkg/engine/datasets"
	"kwil/pkg/engine/models"
	"kwil/pkg/log"
	"kwil/pkg/sql/driver"
	"math/big"
	"os"
	"strings"

	"go.uber.org/zap"
)

// the master is used to connect to the master sqlite database.
// it can find other databases, and it can create new databases.
type Engine struct {
	conn     *driver.Connection
	path     string
	log      log.Logger
	datasets map[string]*datasets.Dataset
}

// Open opens the master database and loads all datasets into memory.
func Open(opts ...MasterOpt) (*Engine, error) {
	e := &Engine{
		log:      log.NewNoOp(),
		path:     DefaultPath,
		datasets: make(map[string]*datasets.Dataset),
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
	for _, dataset := range e.datasets {
		err := dataset.Close()
		if err != nil {
			e.log.Error("failed to close dataset", zap.String("dbid", dataset.DBID), zap.Error(err))
		}
	}

	return e.conn.Close()
}

// createDataset creates a new dataset in the master database, on disk, and in memory.
func (e *Engine) createDataset(owner, name string) error {
	if _, ok := e.datasets[name]; ok {
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
	err = dataset.Clear()
	if err != nil {
		return fmt.Errorf("failed to wipe dataset: %w", err)
	}

	e.datasets[dataset.DBID] = dataset

	return nil
}

// deleteDataset deletes a dataset from the master database, disk, and memory.
// it starts with memory, then disk, then master.
// the order is important because if it fails between disk and master,
// it will catch it when it runs "validateRegisteredDatasets"
func (e *Engine) deleteDataset(dbid string) error {
	delete(e.datasets, dbid)

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

		dataset, err := datasets.OpenDataset(owner, name, e.path)
		if err != nil {
			return fmt.Errorf("failed to open dataset of name: %s, owner: %s: %w", name, owner, err)
		}

		e.datasets[dataset.DBID] = dataset

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

func (e *Engine) Deploy(schema *models.Dataset) error {
	dbid := models.GenerateSchemaId(schema.Owner, schema.Name)
	_, ok := e.datasets[dbid]
	if ok {
		return fmt.Errorf("dataset already exists")
	}

	err := e.createDataset(schema.Owner, schema.Name)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	dataset := e.datasets[dbid]

	err = dataset.ApplySchema(schema)
	if err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	return nil
}

func (e *Engine) DropDataset(dbid string) error {
	ds, ok := e.datasets[dbid]
	if !ok {
		return fmt.Errorf("dataset does not exist")
	}
	defer ds.Close()

	err := ds.Wipe()
	if err != nil {
		// we don't want to return, we can still delete from disk
		e.log.Warn("failed to wipe dataset", zap.Error(err))
	}

	err = e.deleteDataset(dbid)
	if err != nil {
		return fmt.Errorf("failed to delete dataset: %w", err)
	}

	return nil
}

var (
	deployPrice = big.NewInt(1000000000000000000)
	dropPrice   = big.NewInt(10000000000000)
)

func (e *Engine) GetDeployPrice(schema *models.Dataset) (*big.Int, error) {
	return deployPrice, nil
}

func (e *Engine) GetDropPrice(dbid string) (*big.Int, error) {
	return dropPrice, nil
}

func (e *Engine) ListDatabases(owner string) ([]string, error) {
	dbs := make([]string, 0)
	err := e.conn.Query(sqlListDatabasesByOwner, func(stmt *driver.Statement) error {
		dbs = append(dbs, stmt.GetText("name"))
		return nil
	},
		map[string]interface{}{
			"$owner": owner,
		})
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

func (e *Engine) GetDataset(dbid string) (*datasets.Dataset, error) {
	ds, ok := e.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("dataset does not exist")
	}

	return ds, nil
}
