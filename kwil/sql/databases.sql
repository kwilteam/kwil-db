-- name: CreateDatabase :exec
INSERT INTO
    databases (db_name, db_owner)
VALUES
    ($1, (SELECT id FROM accounts WHERE account_address = $2));
    
-- name: DropDatabase :exec
DELETE FROM
    databases
WHERE
    db_name = $1
    AND db_owner = (SELECT id FROM accounts WHERE account_address = $2);

-- name: GetDatabaseId :one
SELECT
    id
FROM
    databases
WHERE
    db_name = $1
    AND db_owner = (SELECT id FROM accounts WHERE account_address = $2);

-- name: ListDatabases :many
SELECT
    db_name,
    account_address
FROM
    databases
    JOIN accounts ON db_owner = accounts.id;
