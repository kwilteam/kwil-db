package engine

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"
	"os"
)

func (e *Engine) addDatasetToMaster(owner, name string) error {
	dbid := models.GenerateSchemaId(owner, name)

	if exists, err := e.ifDatasetExistsInMaster(dbid); err != nil {
		return err
	} else if exists {
		return fmt.Errorf(`dataset "%s" already exists in master database`, dbid)
	}

	return e.conn.ExecuteNamed(sqlCreateDataset, map[string]interface{}{
		"$dbid":  dbid,
		"$owner": owner,
		"$name":  name,
	})
}

func (e *Engine) getDatasetFromMaster(dbid string) (owner string, name string, err error) {
	err = e.conn.QueryNamed(sqlGetDataset, func(stmt *driver.Statement) error {
		owner = stmt.GetText("owner")
		name = stmt.GetText("name")
		return nil
	}, map[string]interface{}{
		"$dbid": dbid,
	})
	if err != nil {
		return "", "", fmt.Errorf(`failed to get dataset "%s" from master database: %w`, dbid, err)
	}

	return owner, name, nil
}

func (e *Engine) ifDatasetExistsInMaster(dbid string) (bool, error) {
	owner, _, err := e.getDatasetFromMaster(dbid)
	if err != nil {
		return false, err
	}

	return owner != "", nil
}

// deleteDatasetFromMaster removes a dataset from the master database.
// it does not delete the dataset from disk.
func (e *Engine) deleteDatasetFromMaster(dbid string) error {
	err := e.conn.ExecuteNamed(sqlDeleteDataset, map[string]interface{}{
		"$dbid": dbid,
	})
	if err != nil {
		return fmt.Errorf(`failed to delete dataset "%s" from master database: %w`, dbid, err)
	}

	return nil
}

// deleteDatasetFromDisk removes a dataset from disk.
// it does not delete the dataset from the master database.
func (e *Engine) deleteDatasetFromDisk(dbid string) error {
	err := os.Remove(e.path + dbid + fileSuffix)
	if err != nil {
		return fmt.Errorf(`failed to delete dataset "%s" from disk: %w`, dbid, err)
	}

	return nil
}
