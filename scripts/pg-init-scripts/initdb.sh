#!/bin/bash
set -e
set -u
function create_user_and_database() {
    local database=$1
    echo "  Creating user and database '$database'"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
        CREATE USER $database;
        CREATE DATABASE $database owner = $database;
        GRANT ALL PRIVILEGES ON DATABASE $database TO $database;
EOSQL
}
function setup_master_db() {
  local database=$1
  echo "  Setting up master database '$database'"
  psql -v ON_ERROR_STOP=1 --username "$database" -d "$database" <<-EOSQL
CREATE TABLE IF NOT EXISTS distributed_locks (
  name CHARACTER VARYING(255) PRIMARY KEY,
  record_version_number BIGINT,
  data BYTEA,
  owner CHARACTER VARYING(255)
);
CREATE SEQUENCE IF NOT EXISTS distributed_locks_rvn OWNED BY distributed_locks.record_version_number;
CREATE TABLE IF NOT EXISTS wallet_info (
  wallet_info_id SERIAL PRIMARY KEY,
  wallet VARCHAR(44) NOT NULL UNIQUE,
  db_connection_url VARCHAR(1000) NOT NULL UNIQUE
 );
  -- NETWORK_METADATA
CREATE TABLE IF NOT EXISTS accounts (
    id serial PRIMARY KEY,
    account_address text NOT NULL,
    balance numeric(78) NOT NULL,
    spent numeric(78) NOT NULL DEFAULT 0,
    nonce bigint NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS deposits (
	id SERIAL PRIMARY KEY,
	tx_hash VARCHAR(66) NOT NULL UNIQUE,
	account_address VARCHAR(44) NOT NULL,
	amount NUMERIC(78) NOT NULL,
	height BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS deposit_height ON deposits(height);

CREATE TABLE IF NOT EXISTS withdrawals (
	id SERIAL PRIMARY KEY,
	correlation_id VARCHAR(10) NOT NULL UNIQUE,
	account_id INTEGER NOT NULL REFERENCES accounts(id),
	amount NUMERIC(78) NOT NULL,
	fee NUMERIC(78) NOT NULL,
	expiry BIGINT NOT NULL,
	tx_hash VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS expiration ON withdrawals(expiry);

-- the height table is meant to be a key value store for the current height
CREATE TABLE IF NOT EXISTS chains (
	id INTEGER PRIMARY KEY,
	chain VARCHAR(20) NOT NULL UNIQUE,
	height BIGINT NOT NULL
);

-- chain ID's do matter, so we make an integer instead of serial.
-- for example, ETHis 1, GOERLI is 2, etc.

-- tables for schemas:
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
EOSQL
}
create_user_and_database "kwil"
setup_master_db "kwil"