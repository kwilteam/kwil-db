-- name: CreateColumn :exec
INSERT INTO
    COLUMNS (table_id, column_name, column_type)
VALUES
    (
        $1,
        $2,
        $3
    );

-- name: CreateAttribute :exec
INSERT INTO
    attributes (column_id, attribute_type, attribute_value)
VALUES
    (
        $1,
        $2,
        $3
    );

-- name: GetColumnId :one
SELECT
    id
FROM
    columns
WHERE
    table_id = $1
    AND column_name = $2;

-- name: GetColumns :many
SELECT
    column_name,
    column_type,
    id
FROM
    columns
WHERE
    table_id = $1;

-- name: GetAttributes :many
SELECT
    attribute_type,
    attribute_value
FROM
    attributes
WHERE
    column_id = $1;