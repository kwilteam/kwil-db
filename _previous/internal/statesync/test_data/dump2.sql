

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

CREATE SCHEMA kwild_accts;


ALTER SCHEMA kwild_accts OWNER TO kwild;


CREATE SCHEMA kwild_voting;


ALTER SCHEMA kwild_voting OWNER TO kwild;

CREATE TABLE kwild_accts.accounts (
    identifier bytea NOT NULL,
    balance text NOT NULL,
    nonce bigint NOT NULL
);


ALTER TABLE kwild_accts.accounts OWNER TO kwild;

CREATE TABLE kwild_voting.resolution_types (
    id bytea NOT NULL,
    name text NOT NULL
);


ALTER TABLE kwild_voting.resolution_types OWNER TO kwild;

CREATE TABLE kwild_voting.voters (
    id bytea NOT NULL,
    name bytea NOT NULL,
    power bigint NOT NULL,
    CONSTRAINT voters_power_check CHECK ((power > 0))
);


ALTER TABLE kwild_voting.voters OWNER TO kwild;


CREATE TABLE kwild_voting.votes (
    resolution_id bytea NOT NULL,
    voter_id bytea NOT NULL
);


ALTER TABLE kwild_voting.votes OWNER TO kwild;


COPY kwild_accts.accounts (identifier, balance, nonce) FROM stdin;
\\xc89d42189f0450c2b2c3c61f58ec5d628176a1e7	0	2
\\xc89dfdg89f04gdt2b2c3c61f58ecfsfsd176a1e7	0	6
\.


COPY kwild_voting.resolution_types (id, name) FROM stdin;
\\xb8272e36a4af5f9da2defaf125a2cfd9	credit_account
\\xa05a316cc7f7557dab180c69acdaecdd	validator_join
\\x2ce0f30017fd500b8c68a72a20fed7bd	validator_remove
\.


COPY kwild_voting.voters (id, name, power) FROM stdin;
\\x8303e83522e158a49da90a29f28bb903	\\x56b7a59cc6373df6aeefc6b21f446d25f5dcc9b9cfb4be7f7d00c9ef94672c83	1
\.



COPY kwild_voting.votes (resolution_id, voter_id) FROM stdin;
\.

