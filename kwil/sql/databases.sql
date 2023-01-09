-- name: CreateDatabase :exec
INSERT INTO
    databases (db_name, db_owner)
VALUES
    ($1, $2);
    
-- name: DropDatabase :exec
DELETE FROM
    databases
WHERE
    db_name = $1
    AND db_owner = $2;