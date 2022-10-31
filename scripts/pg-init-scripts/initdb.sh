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
      CREATE TABLE IF NOT EXISTS users (
        id serial PRIMARY KEY,
        api_key varchar(64) NULL
      );

      CREATE TABLE IF NOT EXISTS wallets (
        wallet_id serial PRIMARY KEY,
        wallet varchar(44),
        balance varchar(20),
        spent varchar(20) DEFAULT '0'
      );

      CREATE TABLE IF NOT EXISTS deposits (
        deposit_id serial PRIMARY KEY,
        txid varchar(64) UNIQUE,
        wallet varchar(44),
        amount varchar(20),
        height INTEGER
      );

      CREATE TABLE IF NOT EXISTS withdrawals (
        withdrawal_id serial PRIMARY KEY,
        nonce varchar(10) UNIQUE,
        wallet varchar(44),
        amount varchar(20),
        fee varchar(20),
        expiry INTEGER,
        wallet_id INTEGER,
        FOREIGN KEY(wallet_id) REFERENCES deposits(deposit_id)
      );
      CREATE INDEX withdrawals_expiry_idx ON withdrawals(expiry);
EOSQL
}

create_user_and_database "kwil"
setup_master_db "kwil"