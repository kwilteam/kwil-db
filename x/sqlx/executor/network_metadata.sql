CREATE TABLE IF NOT EXISTS public.wallets (
		id SERIAL PRIMARY KEY,
		wallet VARCHAR(44) NOT NULL UNIQUE,
		balance NUMERIC(78) DEFAULT '0',
		spent NUMERIC(78) DEFAULT '0'
	);

	CREATE TABLE IF NOT EXISTS public.deposits (
		id SERIAL PRIMARY KEY,
		txid VARCHAR(64) NOT NULL UNIQUE,
		wallet VARCHAR(44) NOT NULL,
		amount NUMERIC(78) NOT NULL,
		height BIGINT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS deposit_height ON public.deposits(height);

	CREATE TABLE IF NOT EXISTS public.withdrawals (
		id SERIAL PRIMARY KEY,
		correlation_id VARCHAR(10) NOT NULL UNIQUE,
		wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
		amount NUMERIC(78),
		fee NUMERIC(78),
		expiry BIGINT NOT NULL,
		tx VARCHAR(64)
	);

    ALTER TABLE public.withdrawals ADD CONSTRAINT withdrawals_wallet_id_fkey FOREIGN KEY (wallet_id) REFERENCES public.wallets(id);

	CREATE INDEX IF NOT EXISTS expiration_ind ON public.withdrawals(expiry);

	-- the height table is meant to be a key value store for the current height
	CREATE TABLE IF NOT EXISTS public.height (
		height BIGINT PRIMARY KEY
	);

CREATE TABLE IF NOT EXISTS public.databases(
    id SERIAL PRIMARY KEY,
    dbs_name VARCHAR(16) NOT NULL UNIQUE,
    database_owner INTEGER NOT NULL,
    default_role INTEGER NOT NULL
);

ALTER TABLE public.databases ADD CONSTRAINT databases_owner_fkey FOREIGN KEY (database_owner) REFERENCES public.wallets(id);

-- get_roles_by_wallet returns the roles for a given wallet
-- it will join the wallets table to the _wallets_roles table, and then to the roles table to return the roles
CREATE OR REPLACE FUNCTION get_roles_by_wallet(wlt varchar(44), dbs varchar(63)) RETURNS SETOF varchar(32) AS $$
    BEGIN
        RETURN QUERY

        EXECUTE 'SELECT role_name FROM ' || dbs || '._roles WHERE id IN (SELECT role_id FROM ' || dbs || '._wallet_roles WHERE wallet_id = (SELECT id FROM public.wallets WHERE wallet = $1))' USING wlt;
    END;
$$ LANGUAGE plpgsql;

-- get_queries_by_role returns the queries for a given role
-- it will join the roles table to the _roles_queries table, and then to the queries table to return the queries
CREATE OR REPLACE FUNCTION get_queries_by_role(role varchar(32), dbs varchar(63)) RETURNS SETOF varchar(32) AS $$
    BEGIN
        RETURN QUERY

        EXECUTE 'SELECT query_name FROM ' || dbs || '._queries WHERE id IN (SELECT query_id FROM ' || dbs || '._roles_queries WHERE role_id = (SELECT id FROM ' || dbs || '._roles WHERE role_name = $1))' USING role;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_role(dbs varchar(63), new_role varchar(32)) RETURNS void AS $$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._roles (role_name) VALUES ($1)' USING new_role;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_query(dbs varchar(63), new_query varchar(32), query_text bytea) RETURNS void AS $$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._queries (query_name, query) VALUES ($1, $2)' USING new_query, query_text;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION add_query_permission(dbs varchar(63), role varchar(32), query varchar(32)) RETURNS void AS $$
    BEGIN
        EXECUTE 'INSERT INTO ' || dbs || '._roles_queries (role_id, query_id) VALUES ((SELECT id FROM ' || dbs || '._roles WHERE role_name = $1), (SELECT id FROM ' || dbs || '._queries WHERE query_name = $2))' USING role, query;
    END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION new_db(nm varchar(61)) RETURNS void AS $$
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
)
';

END;
$$ LANGUAGE plpgsql;
