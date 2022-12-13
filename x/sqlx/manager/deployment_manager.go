package manager

import (
	"context"
	"fmt"
	spec "kwil/x/sqlx"
	"kwil/x/sqlx/models"
	"kwil/x/sqlx/sqlclient"
)

type DeploymentManager interface {
	NewDB(ctx context.Context, nm, owner string, dbBytes []byte) error
	AddRole(ctx context.Context, dbs string, newRole string) error
	AddQuery(ctx context.Context, dbs string, newQuery string, queryText []byte) error
	AddQueryPermission(ctx context.Context, dbs string, role string, query string) error
	DBExists(ctx context.Context, dbs string) (bool, error)
	Delete(ctx context.Context, dbs string) error
	SetDefaultRole(ctx context.Context, dbs string, role string) error
	Store(ctx context.Context, db *models.Database) error
	Deploy(ctx context.Context, db *models.Database) error
}

type deploymentManager struct {
	cache  Cache
	client *sqlclient.DB
}

func NewDeploymentManager(cache Cache, client *sqlclient.DB) *deploymentManager {
	return &deploymentManager{
		cache:  cache,
		client: client,
	}
}

// Deploy will deploy
func (m *deploymentManager) Deploy(ctx context.Context, db *models.Database) error {

	err := db.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate ddl: %w", err)
	}

	// check if the schema exists
	schemaName := db.GetSchemaName()
	exists, err := m.DBExists(ctx, schemaName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("schema %s already exists", schemaName)
	}

	err = m.Store(ctx, db)
	if err != nil {
		return m.revertDeployment(ctx, schemaName, err)
	}

	return nil
}

// AddRole adds a new role to the _roles table
func (m *deploymentManager) AddRole(ctx context.Context, dbs string, newRole string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_role($1, $2)`, dbs, newRole)
	return err
}

func (m *deploymentManager) AddQuery(ctx context.Context, dbs string, newQuery string, queryText []byte) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_query($1, $2, $3)`, dbs, newQuery, queryText)
	return err
}

// AddQueryPermission adds a query permission for a given role
func (m *deploymentManager) AddQueryPermission(ctx context.Context, dbs string, role string, query string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_query_permission($1, $2, $3)`, dbs, role, query)
	return err
}

// NewDB creates a new schema with that name, as well as creates the metadata
func (m *deploymentManager) NewDB(ctx context.Context, nm, owner string, dbBytes []byte) error {
	// nm to lowercase

	_, err := m.client.ExecContext(ctx, `SELECT new_db($1, $2, $3)`, nm, owner, dbBytes)
	return err
}

func (m *deploymentManager) DBExists(ctx context.Context, name string) (bool, error) {
	err := models.CheckName(name, spec.SCHEMA)
	if err != nil {
		return false, err
	}

	var exists bool
	err = m.client.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM public.databases WHERE dbs_name = $1)", name).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// DeleteDB deletes a schema and all of its metadata
func (m *deploymentManager) Delete(ctx context.Context, name string) error {
	err := models.CheckName(name, spec.SCHEMA)
	if err != nil {
		return err
	}

	_, err = m.client.ExecContext(ctx, `select delete_database($1);`, name)
	return err
}

func (m *deploymentManager) SetDefaultRole(ctx context.Context, dbs string, role string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM set_default_role($1, $2)`, dbs, role)
	return err
}

func (m *deploymentManager) Cache(db *models.Database) error {
	return m.cache.Store(db)
}

func (m *deploymentManager) Store(ctx context.Context, db *models.Database) error {
	// create the schema
	schemaName := db.GetSchemaName()
	bts, err := db.EncodeGOB()
	if err != nil {
		return err
	}
	err = m.NewDB(ctx, schemaName, db.Owner, bts)
	if err != nil {
		return fmt.Errorf("failed to create schema. Err: %s: %w", schemaName, err)
	}

	// store ddl
	ddl, err := db.GenerateDDL()
	if err != nil {
		return err
	}

	_, err = m.client.ExecContext(ctx, ddl)
	if err != nil {
		return fmt.Errorf("failed to execute generated ddl: %w", err)
	}

	// set default role
	err = m.SetDefaultRole(ctx, schemaName, db.DefaultRole)
	if err != nil {
		return fmt.Errorf("failed to set default role: %w", err)
	}

	// sync cache
	return m.Cache(db)
}

// revertDeployment attempts to undo a deployment that has failed midway
func (m *deploymentManager) revertDeployment(ctx context.Context, schemaName string, err error) error {
	// delete the schema
	err2 := m.Delete(ctx, schemaName)
	if err2 != nil {
		return fmt.Errorf("failed to delete schema when reverting %s: %s.  Original error: %w", schemaName, err2, err)
	}

	return err
}
