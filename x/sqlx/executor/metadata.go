package executor

import "context"

type MetadataClient interface {
	GetRolesByWallet(wlt string, dbs string) ([]string, error)
	GetQueriesByRole(role string, dbs string) ([]string, error)
	AddRole(dbs string, newRole string) error
	AddQuery(dbs string, newQuery string, queryText []byte) error
	AddQueryPermission(dbs string, role string, query string) error
	NewDB(nm string) error
}

// GetRolesByWallet returns the roles for a given wallet
func (c *client) GetRolesByWallet(ctx context.Context, wlt string, dbs string) ([]string, error) {
	var roles []string

	rows, err := c.db.QueryContext(ctx, `SELECT * FROM get_roles_by_wallet($1, $2)`, wlt, dbs)
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
func (c *client) GetQueriesByRole(ctx context.Context, role string, dbs string) ([]string, error) {
	var queries []string

	rows, err := c.db.QueryContext(ctx, `SELECT * FROM get_queries_by_role($1, $2)`, role, dbs)
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
func (c *client) AddRole(ctx context.Context, dbs string, newRole string) error {
	_, err := c.db.ExecContext(ctx, `SELECT * FROM add_role($1, $2)`, dbs, newRole)
	return err
}

func (c *client) AddQuery(ctx context.Context, dbs string, newQuery string, queryText []byte) error {
	_, err := c.db.ExecContext(ctx, `SELECT * FROM add_query($1, $2, $3)`, dbs, newQuery, queryText)
	return err
}

// AddQueryPermission adds a query permission for a given role
func (c *client) AddQueryPermission(ctx context.Context, dbs string, role string, query string) error {
	_, err := c.db.ExecContext(ctx, `SELECT * FROM add_query_permission($1, $2, $3)`, dbs, role, query)
	return err
}

func (c *client) NewDB(ctx context.Context, nm string) error {
	_, err := c.db.ExecContext(ctx, `SELECT new_db($1)`, nm)
	return err
}

/*
func (c *client) getRolesByWallet(ctx context.Context, wallet, schema string) ([]string, error) {
	res, err := c.db.QueryContext(ctx, "SELECT * FROM get_roles_by_wallet($1, $2)", wallet, schema)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var roles []string
	var role string
	for res.Next() {
		if err := res.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}
*/

/*
func newSchemaScript(name string) (string, error) {
	if len(name) == 0 {
		return "", fmt.Errorf("invalid schema name: %s", name)
	}
	if len(name) > 63 {
		return "", fmt.Errorf("database name too long: %s", name)
	}

	nms := make([]any, SLENGTH)
	for i := range nms {
		nms[i] = name
	}

	// this seems super gross but it will work for now
	return fmt.Sprintf(
		`
-- This is purely used as an example.  The actual script will be kept in a string

CREATE SCHEMA IF NOT EXISTS %s;

CREATE TABLE IF NOT EXISTS %s._queries(
    id SERIAL PRIMARY KEY,
    query_name VARCHAR(32) NOT NULL UNIQUE,
    query BYTEA NOT NULL,
    database_id INTEGER NOT NULL
);
ALTER TABLE %s._queries ADD CONSTRAINT _queries_database_id_fkey FOREIGN KEY (database_id) REFERENCES public.databases(id);

CREATE TABLE IF NOT EXISTS %s._roles(
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(32) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS %s._wallet_roles(
    wallet_id INTEGER NOT NULL REFERENCES _wallets(wallet_id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES _roles(role_id) ON DELETE CASCADE,
    PRIMARY KEY (wallet_id, role_id)
);

ALTER TABLE %s._wallet_roles ADD CONSTRAINT wallet_roles_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id) ON DELETE CASCADE;
ALTER TABLE %s._wallet_roles ADD CONSTRAINT wallet_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES %s._roles(id) ON DELETE CASCADE;

CREATE TABLE IF NOT EXISTS %s._roles_queries(
    role_id INTEGER NOT NULL,
    query_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, query_id)
);

ALTER TABLE %s._roles_queries ADD CONSTRAINT roles_queries_role_id_fkey FOREIGN KEY (role_id) REFERENCES %s._roles(id) ON DELETE CASCADE;
ALTER TABLE %s._roles_queries ADD CONSTRAINT roles_queries_query_id_fkey FOREIGN KEY (query_id) REFERENCES %s._queries(id) ON DELETE CASCADE;

`, nms...), nil
}

const (
	SLENGTH = 13
)
*/
