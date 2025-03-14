-- WARNING this file should be prefixed with a "SET CURRENT NAMESPACE TO" command

CREATE TABLE pending_oracle_data (
    height int primary key,
    previous_height int not null default -1,
    data bytea not null
)

CREATE TABLE pending_oracle_last_processed_height (
    height int primary key
)