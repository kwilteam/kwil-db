CREATE TABLE IF NOT EXISTS wallets (
	wallet_id SERIAL PRIMARY KEY,
	wallet VARCHAR(44) NOT NULL UNIQUE,
	balance NUMERIC(78) DEFAULT '0' NOT NULL,
	spent NUMERIC(78) DEFAULT '0' NOT NULL
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
CREATE TABLE IF NOT EXISTS height (height BIGINT PRIMARY KEY);

