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

CREATE TABLE IF NOT EXISTS wallets (
  wallet_id SERIAL PRIMARY KEY,
  wallet VARCHAR(44) NOT NULL UNIQUE,
  balance NUMERIC(78),
  spent NUMERIC(78) DEFAULT '0'
 );

 CREATE TABLE IF NOT EXISTS deposits (
  deposit_id SERIAL PRIMARY KEY,
  txid VARCHAR(64) NOT NULL UNIQUE,
  wallet VARCHAR(44) NOT NULL,
  amount NUMERIC(78),
  height BIGINT
 );

 CREATE INDEX IF NOT EXISTS deposit_height ON deposits(height);

 CREATE TABLE IF NOT EXISTS withdrawals (
  withdrawal_id SERIAL PRIMARY KEY,
  correlation_id VARCHAR(10) NOT NULL UNIQUE,
  wallet_id INTEGER,
  amount NUMERIC(78),
  fee NUMERIC(78),
  expiry BIGINT,
  tx VARCHAR(64),
  FOREIGN KEY(wallet_id) REFERENCES wallets(wallet_id)
 );

 CREATE INDEX IF NOT EXISTS expiration ON withdrawals(expiry);

 -- the height table is meant to be a key value store for the current height
 CREATE TABLE IF NOT EXISTS height (
  height BIGINT PRIMARY KEY
 );

 CREATE OR REPLACE FUNCTION set_height(h BIGINT) RETURNS VOID AS \$\$
 BEGIN
  IF EXISTS (SELECT 1 FROM height) THEN
   UPDATE height SET height = h;
  ELSE
   INSERT INTO height VALUES (h);
  END IF;
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION get_height() RETURNS BIGINT AS \$\$
 BEGIN
  -- get the height from the height table 
  RETURN (select coalesce((SELECT height FROM height), 0));
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION deposit(t VARCHAR(64), w VARCHAR(44), a NUMERIC(78), h BIGINT) RETURNS VOID AS \$\$
 BEGIN
  INSERT INTO deposits (txid, wallet, amount, height) VALUES (t, w, a, h);
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION spend_money(addr VARCHAR(44), amt numeric(78)) RETURNS void AS \$\$
  DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr); 
  BEGIN
   -- TODO: Add sequence ID
   IF oldBal - amt < 0 THEN RAISE EXCEPTION 'not enough balance'; END IF;
   UPDATE wallets SET balance = oldBal-amt, spent = (SELECT spent FROM wallets WHERE wallet = addr) + amt WHERE wallet = addr;
  
  END; 
  \$\$ LANGUAGE plpgsql;
 
 CREATE OR REPLACE FUNCTION commit_deposits(ht BIGINT) RETURNS void AS \$\$
  DECLARE d deposits%ROWTYPE;
  BEGIN
   -- loop through all deposits at or below the height.  If the wallet already has a balance, it will add the deposit to the balance
   -- If the wallet does not have a balance, it will create a new balance.  Then it will delete the deposit
   FOR d IN SELECT * FROM deposits WHERE height <= ht LOOP
    INSERT INTO wallets (wallet, balance) VALUES (d.wallet, d.amount) ON CONFLICT (wallet) DO UPDATE SET balance = wallets.balance + d.amount;
    DELETE FROM deposits WHERE deposit_id = d.deposit_id;
   END LOOP;

  END;
  \$\$ LANGUAGE plpgsql;
 
 CREATE OR REPLACE FUNCTION start_withdrawal(addr varchar(44), cid varchar(10), amt numeric(78), expiry BIGINT) RETURNS table (ret_addr varchar(44), ret_cid varchar(10), ret_amount numeric(78), ret_fee numeric(78), ret_expiration BiGINT) AS \$\$
  DECLARE wid integer = (SELECT wallet_id FROM wallets WHERE wallet = addr);
  DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr);
  DECLARE spent NUMERIC(78) = (SELECT spent FROM wallets WHERE wallet = addr);

  BEGIN
   -- If oldBal > amt, then make amt = oldBal and set oldBal to 0
   IF oldBal - amt < 0 THEN amt = oldBal; oldBal = 0; ELSE oldBal = oldBal - amt; END IF;
   -- If both spent and amt are 0, then raise an exception
   IF spent + amt = 0 THEN RAISE EXCEPTION 'cannot withdraw with 0 balance and spent'; END IF;

   -- Now we insert into withdrawals
   INSERT INTO withdrawals (correlation_id, wallet_id, amount, fee, expiry) VALUES (cid, wid, amt, spent, expiry);
   -- Now we update the wallets table
   UPDATE wallets SET balance = oldBal, spent = 0 WHERE wallet_id = wid;

   RETURN QUERY SELECT addr, cid, amt, spent, expiry;
  END;

  \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION add_tx(cid VARCHAR(10), txh VARCHAR(64)) RETURNS void AS \$\$ 
  BEGIN
   UPDATE withdrawals SET tx = txh WHERE correlation_id = cid;
  END;
  \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION remove_balance(addr varchar(44), amount numeric(78)) RETURNS void AS \$\$  
-- subtract the balance from the wallet
  DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr);

  BEGIN
   IF oldBal - amount < 0 THEN RAISE EXCEPTION 'not enough balance'; END IF;
   UPDATE wallets SET balance = oldBal - amount WHERE wallet = addr;
  END;
  \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION expire(height BIGINT) RETURNS VOID AS \$\$
  DECLARE w withdrawals%ROWTYPE;
  BEGIN
   -- loop through all withdrawals at or below the height.  Transfer the amount back to the balance, and the fee back to the spent, and delete the withdrawal
 
   FOR w IN SELECT * FROM withdrawals WHERE expiry <= height LOOP
    UPDATE wallets SET balance = balance + w.amount, spent = spent + w.fee WHERE wallet_id = w.wallet_id;
    DELETE FROM withdrawals WHERE withdrawal_id = w.withdrawal_id;
   END LOOP;
 
   END;
  \$\$ LANGUAGE plpgsql;
 
 -- finish_withdrawal will rewturn a boolean for whether or not there was a withdrawal with that correlation_id
 CREATE OR REPLACE FUNCTION finish_withdrawal(n VARCHAR(10)) RETURNS BOOLEAN AS \$\$
 BEGIN
  -- delete the withdrawal with the given cid
  DELETE FROM withdrawals WHERE correlation_id = n;
  -- return true if there was a withdrawal with that cid
  RETURN FOUND;
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION commit_height(ht BIGINT) RETURNS VOID AS \$\$
 BEGIN
  PERFORM commit_deposits(ht);
  PERFORM expire(ht);
  PERFORM set_height(ht+1);
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION get_balance(addr VARCHAR(44)) RETURNS NUMERIC(78) AS \$\$
 BEGIN
  RETURN (SELECT balance FROM wallets WHERE wallet = addr);
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION get_spent(addr VARCHAR(44)) RETURNS NUMERIC(78) AS \$\$
 BEGIN
  RETURN (SELECT spent FROM wallets WHERE wallet = addr);
 END; 
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION get_balance_and_spent(addr VARCHAR(44)) RETURNS TABLE (balance NUMERIC(78), spent NUMERIC(78)) AS \$\$
 BEGIN
  RETURN QUERY SELECT balance, spent FROM wallets WHERE wallet = addr;
 END;
 \$\$ LANGUAGE plpgsql;

 CREATE OR REPLACE FUNCTION get_withdrawals_addr(addr VARCHAR(44)) RETURNS TABLE (n VARCHAR(10), a NUMERIC(78), s NUMERIC(78), e BIGINT, t varchar(64)) AS \$\$
 BEGIN
  RETURN QUERY SELECT correlation_id, amount, fee, expiry, tx FROM withdrawals WHERE wallet_id = (SELECT wallet_id FROM wallets WHERE wallet = addr);
 END;
 \$\$ LANGUAGE plpgsql;
 
 CREATE OR REPLACE FUNCTION get_all_withdrawals(h BIGINT) RETURNS TABLE (n VARCHAR(10), a NUMERIC(78), s NUMERIC(78), e BIGINT, w varchar(44)) AS \$\$
 -- must return cid, amount, fee, expiry, and wallet (based on wallet_id)
 BEGIN
  -- Get withdrawals at or before height h and join with wallets to get the wallet
  -- Return the cid, amount, fee, expiry, and wallet
  RETURN QUERY SELECT correlation_id, amount, fee, expiry, wallet FROM withdrawals JOIN wallets ON withdrawals.wallet_id = wallets.wallet_id WHERE expiry <= h;
 END;
 \$\$ LANGUAGE plpgsql;


  -- NETWORK_METADATA


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
    dbs_name VARCHAR(63) NOT NULL UNIQUE,
    database_owner INTEGER NOT NULL,
    default_role INTEGER
);

CREATE TABLE IF NOT EXISTS public.database_schemas(
    id SERIAL PRIMARY KEY,
    dbs_id INTEGER NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    db_schema BYTEA
);

ALTER TABLE public.databases ADD CONSTRAINT databases_owner_fkey FOREIGN KEY (database_owner) REFERENCES public.wallets(id);

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