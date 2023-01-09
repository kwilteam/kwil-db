-- name: CreateTable :exec
INSERT INTO
    tables (db_id, table_name)
VALUES
    ((SELECT id FROM databases WHERE db_name = $1), $2);

-- name: ListTables :many
SELECT
    table_name
FROM
    tables
WHERE
    db_id = (SELECT id FROM databases WHERE db_name = $1 AND db_owner = $2);