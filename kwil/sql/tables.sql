-- name: CreateTable :exec
INSERT INTO
    tables (db_id, table_name)
VALUES
    ($1, $2);

-- name: ListTables :many
SELECT
    table_name,
    id
FROM
    tables
WHERE
    db_id = $1;

-- name: GetTableId :one
SELECT
    id
FROM
    tables
WHERE
    db_id = $1
    AND table_name = $2;

