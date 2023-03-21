package datasets

import (
	"kwil/pkg/engine/models"
	"kwil/pkg/sql/driver"
)

type Dataset struct {
	conn       *driver.Connection
	Owner      string
	Name       string
	DBID       string
	schema     *models.Dataset
	statements map[string]*PreparedStatement
}

func (d *Dataset) newPreparedStatement(stmt string) error {
	// TODO: implement
	return nil
}

func OpenDataset(owner, name, path string) (*Dataset, error) {
	dbid := models.GenerateSchemaId(owner, name)

	conn, err := driver.OpenConn(dbid,
		driver.WithPath(path),
	)
	if err != nil {
		return nil, err
	}

	err = conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	return &Dataset{
		conn:       conn,
		Owner:      owner,
		Name:       name,
		DBID:       dbid,
		schema:     nil, // TODO: load schema from disk
		statements: nil, // TODO: load prepared statements from disks
	}, nil
}
