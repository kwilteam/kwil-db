CREATE TABLE IF NOT EXISTS wallets (
	id SERIAL PRIMARY KEY,
	wallet VARCHAR(44) NOT NULL UNIQUE,
	balance NUMERIC(78) DEFAULT '0' NOT NULL,
	spent NUMERIC(78) DEFAULT '0' NOT NULL
);

CREATE TABLE IF NOT EXISTS deposits (
	id SERIAL PRIMARY KEY,
	tx_hash VARCHAR(64) NOT NULL UNIQUE,
	wallet VARCHAR(44) NOT NULL,
	amount NUMERIC(78) NOT NULL,
	height BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS deposit_height ON deposits(height);

CREATE TABLE IF NOT EXISTS withdrawals (
	id SERIAL PRIMARY KEY,
	correlation_id VARCHAR(10) NOT NULL UNIQUE,
	wallet_id INTEGER NOT NULL REFERENCES wallets(id),
	amount NUMERIC(78) NOT NULL,
	fee NUMERIC(78) NOT NULL,
	expiry BIGINT NOT NULL,
	tx_hash VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS expiration ON withdrawals(expiry);

-- the height table is meant to be a key value store for the current height
CREATE TABLE IF NOT EXISTS height (height BIGINT PRIMARY KEY);