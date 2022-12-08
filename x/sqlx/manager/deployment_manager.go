package manager

import (
	"context"
	"fmt"
	"kwil/x/sqlx/schema"
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
	SyncCache(ctx context.Context, db *schema.Database, dbs string) error
	Store(ctx context.Context, db *schema.Database) error
	Deploy(ctx context.Context, owner string, ddl []byte) error
}

type deploymentManager struct {
	cache  *SchemaCache
	client *sqlclient.DB
}

func NewDeploymentManager(cache *SchemaCache, client *sqlclient.DB) *deploymentManager {
	return &deploymentManager{
		cache:  cache,
		client: client,
	}
}

// Deploy will deploy
func (m *deploymentManager) Deploy(ctx context.Context, owner string, ddl []byte) error {
	db := &schema.Database{}
	err := db.UnmarshalYAML(ddl)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ddl: %w", err)
	}

	if db.Owner != owner {
		return fmt.Errorf("owner mismatch: %s != %s", db.Owner, owner)
	}

	err = db.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate ddl: %w", err)
	}

	// check if the schema exists
	schemaName := db.SchemaName()
	exists, err := m.DBExists(ctx, schemaName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("schema %s already exists", schemaName)
	}

	err = m.Store(ctx, db)
	if err != nil {
		return m.revertDeployment(ctx, db.SchemaName(), err)
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
	val, err := schema.CheckValidName(name)
	if err != nil {
		return false, err
	}
	if !val {
		return false, fmt.Errorf("invalid schema name: %s", name)
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
	val, err := schema.CheckValidName(name)
	if err != nil {
		return err
	}
	if !val {
		return fmt.Errorf("invalid schema name: %s", name)
	}

	_, err = m.client.ExecContext(ctx, `select delete_database($1);`, name)
	return err
}

func (m *deploymentManager) SetDefaultRole(ctx context.Context, dbs string, role string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM set_default_role($1, $2)`, dbs, role)
	return err
}

func (m *deploymentManager) SyncCache(ctx context.Context, db *schema.Database, dbs string) error {
	return m.cache.Sync(ctx, db, dbs)
}

func (m *deploymentManager) Store(ctx context.Context, db *schema.Database) error {
	// create the schema
	schemaName := db.SchemaName()
	bts, err := db.EncodeGOB()
	if err != nil {
		return err
	}
	err = m.NewDB(ctx, schemaName, db.Owner, bts)
	if err != nil {
		return fmt.Errorf("failed to create schema.  Make sure the owner is registed in the wallets table.  Err: %s: %w", schemaName, err)
	}

	// store ddl
	ddl, err := db.GenerateDDL()
	if err != nil {
		return err
	}

	for _, stmt := range ddl {
		_, err := m.client.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute generated ddl: %w", err)
		}
	}

	// store queries
	queries := db.Queries.GetAll()
	for name, query := range queries {
		executable, err := query.Prepare(db)
		if err != nil {
			return fmt.Errorf("failed to prepare query %s: %w", name, err)
		}

		execBytes, err := executable.Bytes()
		if err != nil {
			return fmt.Errorf("failed to get bytes of query %s: %w", name, err)
		}

		err = m.AddQuery(ctx, schemaName, name, execBytes)
		if err != nil {
			return fmt.Errorf("failed to add query %s: %w", name, err)
		}
	}

	// store roles
	for name, role := range db.Roles {
		err := m.AddRole(ctx, schemaName, name)
		if err != nil {
			return fmt.Errorf("failed to add role: %w", err)
		}
		for _, queryName := range role.Queries {
			err = m.AddQueryPermission(ctx, schemaName, name, queryName)
			if err != nil {
				return fmt.Errorf("failed to add query permission: %w", err)
			}
		}
	}

	// set default role
	err = m.SetDefaultRole(ctx, schemaName, db.DefaultRole)
	if err != nil {
		return fmt.Errorf("failed to set default role: %w", err)
	}

	// sync cache
	return m.SyncCache(ctx, db, schemaName)
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
