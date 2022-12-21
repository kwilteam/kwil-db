package manager

import (
	"context"
	spec "kwil/x/sqlx"
	"kwil/x/sqlx/models"
	"kwil/x/sqlx/sqlclient"
	"strings"
)

type MetadataManager interface {
	GetDefaultRole(ctx context.Context, db string) (string, error)
	GetRolesByWallet(ctx context.Context, wlt string, dbs string) ([]string, error)
	GetQueriesByRole(ctx context.Context, role string, dbs string) ([]string, error)
	ListRoles(ctx context.Context, dbs string) ([]string, error)
	GetMetadataBytes(ctx context.Context, dbs string) ([]byte, error)
	GetAllDatabases(ctx context.Context) ([]string, error)
	ListDatabases(ctx context.Context, owner string) ([]string, error)
}

type metadataManager struct {
	client *sqlclient.DB
}

func NewMetadataManager(client *sqlclient.DB) *metadataManager {
	return &metadataManager{
		client: client,
	}
}

// GetDefaultRole returns the default role for a given database
func (m *metadataManager) GetDefaultRole(ctx context.Context, db string) (string, error) {
	var role string
	err := m.client.QueryRowContext(ctx, `SELECT * FROM get_default_role($1)`, db).Scan(&role)
	return role, err
}

// GetRolesByWallet returns the roles for a given wallet
func (m *metadataManager) GetRolesByWallet(ctx context.Context, wlt string, dbs string) ([]string, error) {
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
func (m *metadataManager) GetQueriesByRole(ctx context.Context, role string, dbs string) ([]string, error) {
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

// ListRoles returns the list of roles
func (m *metadataManager) ListRoles(ctx context.Context, dbs string) ([]string, error) {
	var roles []string

	rows, err := m.client.QueryContext(ctx, `SELECT * FROM list_roles($1)`, dbs)
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

// GetMetadataBytes returns the metadata bytes
func (m *metadataManager) GetMetadataBytes(ctx context.Context, dbs string) ([]byte, error) {
	// since this gets called from different areas, will check name to be safe
	err := models.CheckName(dbs, spec.SCHEMA)
	if err != nil {
		return nil, err
	}

	var metadataBytes []byte
	err = m.client.QueryRowContext(ctx, `SELECT db_schema FROM public.database_schemas WHERE dbs_id = (SELECT id FROM public.databases WHERE dbs_name = $1)`, dbs).Scan(&metadataBytes)
	return metadataBytes, err
}

// GetAllDatabases returns all the databases
func (m *metadataManager) GetAllDatabases(ctx context.Context) ([]string, error) {
	var databases []string

	rows, err := m.client.QueryContext(ctx, `SELECT dbs_name FROM public.databases`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, err
		}
		databases = append(databases, database)
	}

	return databases, nil
}

func (m *metadataManager) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	var databases []string

	rows, err := m.client.QueryContext(ctx, `SELECT dbs_name FROM public.databases WHERE database_owner = $1`, strings.ToLower(owner))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		if err := rows.Scan(&database); err != nil {
			return nil, err
		}
		databases = append(databases, database)
	}

	return databases, nil
}
