CREATE NAMESPACE kwil_erc20_meta;
SET CURRENT NAMESPACE TO kwil_erc20_meta;

-- reward_instances tracks all reward extensions that have been created.
-- it includes reward_instances that have synced data from the chain, as
-- well as reward_instances that have been created but not yet synced.
CREATE TABLE reward_instances (
    -- the following columns are set when the reward is created
    id UUID PRIMARY KEY,
    chain_id TEXT NOT NULL,
    escrow_address BYTEA NOT NULL,
    distribution_period INT8 NOT NULL, -- interval (in seconds)
    synced BOOLEAN NOT NULL DEFAULT FALSE, -- whether the reward has been synced
    active BOOLEAN NOT NULL DEFAULT TRUE, -- whether the reward is active
    -- the following columns are set when the on-chain info
    -- is synced
    erc20_address BYTEA,
    erc20_decimals INT8,
    synced_at INT8, -- the unix timestamp (in seconds) when the reward was synced
    balance NUMERIC(78, 0) NOT NULL DEFAULT 0 CHECK(balance >= 0) -- the total balance owned by the database that can be distributed
);

-- balances tracks the balance of each user in a given reward instance.
CREATE TABLE balances (
    id UUID PRIMARY KEY,
    reward_id UUID NOT NULL REFERENCES reward_instances(id) ON UPDATE CASCADE ON DELETE CASCADE,
    address BYTEA NOT NULL, -- wallet address of the user
    balance NUMERIC(78, 0) NOT NULL DEFAULT 0 CHECK(balance >= 0) -- the balance owned by the user on this network
);

-- epochs holds the epoch information.
-- An epoch is a group of rewards that are issued in a given time/block range.
-- Epochs have 3 states:
-- 1. Created: the epoch is created and the rewards are being distributed
-- 2. Ended: the epoch has ended and the rewards are finalized
-- 3. Confirmed: the epoch has been confirmed on chain
-- Ideally, Kwil would have a unique indexes on this table where the ended_at is null (to enforce only one active epoch at a time),
-- but this requires partial indexes which are not yet supported in Kwil
CREATE TABLE epochs (
	id UUID PRIMARY KEY,
    created_at INT8 NOT NULL, -- kwil block height
    instance_id UUID NOT NULL REFERENCES reward_instances(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	reward_root BYTEA UNIQUE, -- the root of the merkle tree of rewards, it's unique per contract
    ended_at INT8, -- kwil block height
    block_hash BYTEA, -- the hash of the block that is used in merkle tree leaf, which is the last block of the epoch
    confirmed BOOLEAN NOT NULL DEFAULT FALSE -- whether the epoch has been confirmed on chain
);

-- epoch_rewards holds information about the rewards in a given epoch
CREATE TABLE epoch_rewards (
    epoch_id UUID NOT NULL REFERENCES epochs(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    recipient BYTEA NOT NULL,
    amount NUMERIC(78,0) NOT NULL, -- allows uint256
    PRIMARY KEY (epoch_id, recipient)
);