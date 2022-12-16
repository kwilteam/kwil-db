-- name: SetHeight :exec
UPDATE
    height
SET
    height = $1;

-- name: GetHeight :one
SELECT
    height
FROM
    height
LIMIT
    1;