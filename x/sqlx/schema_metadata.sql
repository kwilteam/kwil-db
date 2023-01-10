-- This is purely used as an example.  The actual script will be kept in a string in a function

CREATE TABLE IF NOT EXISTS owner_name._queries(
    id SERIAL PRIMARY KEY,
    query_name VARCHAR(32) NOT NULL UNIQUE,
    query BYTEA
);

CREATE TABLE IF NOT EXISTS owner_name._roles(
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(32) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS owner_name._wallet_roles(
    wallet_id INTEGER NOT NULL REFERENCES public.wallets(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES owner_name._roles(id) ON DELETE CASCADE,
    PRIMARY KEY (wallet_id, role_id)
);

CREATE TABLE IF NOT EXISTS owner_name._roles_queries(
    role_id INTEGER NOT NULL REFERENCES owner_name._roles(id) ON DELETE CASCADE,
    query_id INTEGER NOT NULL REFERENCES owner_name._queries(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, query_id)
);