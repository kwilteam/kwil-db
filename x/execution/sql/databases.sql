-- name: CreateDatabase :exec
INSERT INTO
    databases (db_name, db_owner)
VALUES
    ($1, $2);
