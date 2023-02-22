-- name: GetTableSize :one
SELECT
    pg_relation_size(relid)
FROM
    pg_catalog.pg_statio_user_tables
WHERE
    schemaname = $1
    AND relname = $2;

-- name: GetIndexedColumnCount :one
SELECT
    count(*)
FROM
    pg_catalog.pg_indexes
WHERE
    schemaname = $1
    AND tablename = $2;