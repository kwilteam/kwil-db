package schema_manager

import (
	"context"
	"kwil/x/sqlx/sqlclient"
)

/*
	A schema manager is simply a wrapper around the sqlclient that provides
	an easy interface for creating and managing user-deployed databases and their metadata.
*/

// Manager represents a manager interface
type Manager interface {
	GetRolesByWallet(ctx context.Context, wlt string, dbs string) ([]string, error)
	GetQueriesByRole(ctx context.Context, role string, dbs string) ([]string, error)
	AddRole(ctx context.Context, dbs string, newRole string) error
	AddQuery(ctx context.Context, dbs string, newQuery string, queryText []byte) error
	AddQueryPermission(ctx context.Context, dbs string, role string, query string) error
	NewDB(ctx context.Context, nm string) error
	SchemaExists(ctx context.Context, name string) (bool, error)
	DeleteDB(ctx context.Context, name string) error
	Client() *sqlclient.DB
}

// manager implements the Manager interface
type manager struct {
	// The database client
	client *sqlclient.DB
}

// NewManager returns a new manager
func New(client *sqlclient.DB) *manager {
	return &manager{
		client: client,
	}
}

// GetRolesByWallet returns the roles for a given wallet
func (m *manager) GetRolesByWallet(ctx context.Context, wlt string, dbs string) ([]string, error) {
	var roles []string

	rows, err := m.client.QueryContext(ctx, `SELECT * FROM get_roles_by_wallet($1, $2)`, wlt, dbs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetQueriesByRole returns the queries for a given role
func (m *manager) GetQueriesByRole(ctx context.Context, role string, dbs string) ([]string, error) {
	var queries []string

	rows, err := m.client.QueryContext(ctx, `SELECT * FROM get_queries_by_role($1, $2)`, role, dbs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var query string
		if err := rows.Scan(&query); err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}

	return queries, nil
}

// AddRole adds a new role to the _roles table
func (m *manager) AddRole(ctx context.Context, dbs string, newRole string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_role($1, $2)`, dbs, newRole)
	return err
}

func (m *manager) AddQuery(ctx context.Context, dbs string, newQuery string, queryText []byte) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_query($1, $2, $3)`, dbs, newQuery, queryText)
	return err
}

// AddQueryPermission adds a query permission for a given role
func (m *manager) AddQueryPermission(ctx context.Context, dbs string, role string, query string) error {
	_, err := m.client.ExecContext(ctx, `SELECT * FROM add_query_permission($1, $2, $3)`, dbs, role, query)
	return err
}

// NewDB creates a new schema with that name, as well as creates the metadata
func (m *manager) NewDB(ctx context.Context, nm string) error {
	_, err := m.client.ExecContext(ctx, `SELECT new_db($1)`, nm)
	return err
}

func (m *manager) SchemaExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := m.client.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pg_namespace WHERE nspname = $1)", name).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// DeleteDB deletes a schema and all of its metadata
func (m *manager) DeleteDB(ctx context.Context, name string) error {
	_, err := m.client.ExecContext(ctx, `DROP SCHEMA IF EXISTS $1 CASCADE`, name)
	return err
}

func (m *manager) Client() *sqlclient.DB {
	return m.client
}
