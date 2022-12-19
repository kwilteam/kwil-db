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
CREATE TABLE IF NOT EXISTS public.wallets (
		id SERIAL PRIMARY KEY,
		wallet VARCHAR(44) NOT NULL UNIQUE,
		balance NUMERIC(78) DEFAULT '0',
		spent NUMERIC(78) DEFAULT '0'
	);

CREATE TABLE IF NOT EXISTS public.databases(
    id SERIAL PRIMARY KEY,
    dbs_name VARCHAR(63) NOT NULL UNIQUE,
    database_owner INTEGER NOT NULL,
    default_role INTEGER
);
CREATE TABLE IF NOT EXISTS public.database_schemas(
    id SERIAL PRIMARY KEY,
    dbs_id INTEGER NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    db_schema BYTEA
);
CREATE OR REPLACE FUNCTION get_default_role(dbs varchar(63)) RETURNS SETOF varchar(32) AS \$\$
    BEGIN
        RETURN QUERY
        EXECUTE 'SELECT role_name FROM ' || dbs || '._roles WHERE id = (SELECT default_role FROM public.databases WHERE dbs_name = \$1)' USING dbs;
    END;
\$\$ LANGUAGE plpgsql;
-- get_roles_by_wallet returns the roles for a given wallet
-- it will join the wallets table to the _wallets_roles table, and then to the roles table to return the roles
CREATE OR REPLACE FUNCTION get_roles_by_wallet(wlt varchar(44), dbs varchar(63)) RETURNS SETOF varchar(32) AS \$\$
    BEGIN
        RETURN QUERY
        EXECUTE 'SELECT role_name FROM ' || dbs || '._roles WHERE id IN (SELECT role_id FROM ' || dbs || '._wallet_roles WHERE wallet_id = (SELECT id FROM public.wallets WHERE wallet = \$1))' USING wlt;
    END;
\$\$ LANGUAGE plpgsql;
-- get_queries_by_role returns the queries for a given role
-- it will join the roles table to the _roles_queries table, and then to the queries table to return the queries
CREATE OR REPLACE FUNCTION get_queries_by_role(role varchar(32), dbs varchar(63)) RETURNS SETOF varchar(32) AS \$\$
    BEGIN
        RETURN QUERY
        EXECUTE 'SELECT query_name FROM ' || dbs || '._queries WHERE id IN (SELECT query_id FROM ' || dbs || '._roles_queries WHERE role_id = (SELECT id FROM ' || dbs || '._roles WHERE role_name = \$1))' USING role;
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION list_roles(dbs varchar(63)) RETURNS SETOF varchar(32) AS \$\$
    BEGIN
        RETURN QUERY EXECUTE 'SELECT role_name FROM ' || dbs || '._roles';
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION add_role(dbs varchar(63), new_role varchar(32)) RETURNS void AS \$\$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._roles (role_name) VALUES (\$1)' USING new_role;
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION add_query(dbs varchar(63), new_query varchar(32), query_text bytea) RETURNS void AS \$\$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._queries (query_name, query) VALUES (\$1, \$2)' USING new_query, query_text;
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION add_query_permission(dbs varchar(63), rol varchar(32), query varchar(32)) RETURNS void AS \$\$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._roles_queries (role_id, query_id) VALUES ((SELECT id FROM ' || dbs || '._roles WHERE role_name = \$1), (SELECT id FROM ' || dbs || '._queries WHERE query_name = \$2))' USING rol, query;
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION new_db(nm varchar(61), ownr varchar(44), schma BYTEA) RETURNS void AS \$\$
BEGIN
-- check if schema already exists
IF EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = nm) THEN
    RAISE EXCEPTION 'Schema already exists';
END IF;
    EXECUTE 'CREATE SCHEMA ' || nm;
    EXECUTE 'CREATE TABLE IF NOT EXISTS ' || nm || '._queries(
    id SERIAL PRIMARY KEY,
    query_name VARCHAR(32) NOT NULL UNIQUE,
    query BYTEA
);
CREATE TABLE IF NOT EXISTS ' || nm || '._roles(
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(32) NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS ' || nm || '._wallet_roles(
    wallet_id INTEGER NOT NULL REFERENCES public.wallets(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES ' || nm || '._roles(id) ON DELETE CASCADE,
    PRIMARY KEY (wallet_id, role_id)
);
CREATE TABLE IF NOT EXISTS ' || nm || '._roles_queries(
    role_id INTEGER NOT NULL REFERENCES ' || nm || '._roles(id) ON DELETE CASCADE,
    query_id INTEGER NOT NULL REFERENCES ' || nm || '._queries(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, query_id)
);
INSERT INTO public.databases (dbs_name, database_owner) VALUES (\$1, (SELECT id FROM public.wallets WHERE wallet = \$2));
INSERT INTO public.database_schemas (dbs_id, db_schema) VALUES ((SELECT id FROM public.databases WHERE dbs_name = \$1), \$3);
' USING nm, ownr, schma;
END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION set_default_role(dbs varchar(63), def_role varchar(32)) RETURNS void AS \$\$
    BEGIN
        EXECUTE 'UPDATE public.databases SET default_role = (SELECT id FROM ' || dbs || '._roles WHERE role_name = \$1) WHERE dbs_name = \$2' USING def_role, dbs;
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION get_all_queries (dbs varchar(63)) RETURNS TABLE(q_n varchar(32), qry BYTEA) AS \$\$
    BEGIN
        RETURN QUERY EXECUTE 'SELECT query_name, query FROM ' || dbs || '._queries';
    END;
\$\$ LANGUAGE plpgsql;
CREATE OR REPLACE FUNCTION delete_database(dbs varchar(63)) RETURNS void AS \$\$
    BEGIN
        EXECUTE 'DROP SCHEMA ' || dbs || ' CASCADE';
        EXECUTE 'DELETE FROM public.databases WHERE dbs_name = \$1;' USING dbs;
        EXECUTE 'DELETE FROM public.database_schemas WHERE dbs_id = (SELECT id FROM public.databases WHERE dbs_name = \$1)' USING dbs;
    END;
\$\$ LANGUAGE plpgsql;
EOSQL
}
create_user_and_database "kwil"
setup_master_db "kwil"