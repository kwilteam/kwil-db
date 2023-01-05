CREATE TABLE IF NOT EXISTS wallets (
	id SERIAL PRIMARY KEY,
	wallet VARCHAR(44) NOT NULL UNIQUE,
	balance NUMERIC(78) DEFAULT '0' NOT NULL,
	spent NUMERIC(78) DEFAULT '0' NOT NULL
);

CREATE TABLE IF NOT EXISTS databases (
    id INTEGER PRIMARY KEY,
    db_owner TEXT NOT NULL REFERENCES wallets (wallet) ON DELETE CASCADE,
    db_name TEXT NOT NULL,
    unique (db_owner, db_name)
);

CREATE TABLE IF NOT EXISTS tables (
    id INTEGER PRIMARY KEY,
    db_id INTEGER REFERENCES databases (id) ON DELETE CASCADE,
    table_name TEXT NOT NULL,
    unique (db_id, table_name)
);

CREATE TABLE IF NOT EXISTS columns (
    id INTEGER PRIMARY KEY,
    table_id INTEGER REFERENCES tables (id) ON DELETE CASCADE,
    column_name TEXT NOT NULL,
    column_type INTEGER NOT NULL,
    unique (table_id, column_name)
);

CREATE TABLE IF NOT EXISTS attributes (
    id INTEGER PRIMARY KEY,
    column_id INTEGER REFERENCES columns (id) ON DELETE CASCADE,
    attribute_type INTEGER NOT NULL,
    attribute_value BYTEA,
    unique (column_id, attribute_type)
);

CREATE TABLE IF NOT EXISTS indexes (
    id INTEGER PRIMARY KEY,
    table_id INTEGER NOT NULL REFERENCES tables (id) ON DELETE CASCADE,
    columns TEXT[] NOT NULL,
    index_name TEXT NOT NULL,
    index_type INTEGER NOT NULL,
    unique (table_id, index_name)
);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    role_name TEXT NOT NULL,
    db_id INTEGER REFERENCES databases (id) ON DELETE CASCADE,
    unique (role_name)
);

-- join table for many-to-many relationship between roles and wallets
CREATE TABLE IF NOT EXISTS role_wallets (
    role_id INTEGER REFERENCES roles (id) ON DELETE CASCADE,
    wallet_id INTEGER REFERENCES wallets (id) ON DELETE CASCADE,
    unique (role_id, wallet_id)
);

CREATE TABLE IF NOT EXISTS queries (
    id INTEGER PRIMARY KEY,
    query_name TEXT NOT NULL,
    query BYTEA NOT NULL,
    table_id INTEGER REFERENCES tables (id) ON DELETE CASCADE, 
    unique (query_name)
);

-- join table for many-to-many relationship between roles and queries
CREATE TABLE IF NOT EXISTS role_queries (
    role_id INTEGER REFERENCES roles (id) ON DELETE CASCADE,
    query_id INTEGER REFERENCES queries (id) ON DELETE CASCADE,
    unique (role_id, query_id)
);