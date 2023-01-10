-- name: CreateIndex :exec
INSERT INTO
    INDEXES (table_id, index_name, index_type, COLUMNS)
VALUES
    (
        $1,
        $2,
        $3,
        $4
    );

-- name: GetIndexes :many
SELECT
    index_name,
    index_type,
    columns
FROM
    indexes
WHERE
    table_id = $1;