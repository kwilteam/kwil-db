-- name: CreateQuery :exec
INSERT INTO
    queries (query_name, query, db_id) VALUES
    (
        $1,
        $2,
        $3
    );

-- name: GetQueries :many
SELECT
    query,
    id
FROM
    queries
WHERE
    db_id = $1;