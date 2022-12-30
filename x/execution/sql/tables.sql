-- name: CreateTable :exec
INSERT INTO
    tables (db_id, table_name)
VALUES
    ((SELECT id FROM databases WHERE db_name = $1), $2);

