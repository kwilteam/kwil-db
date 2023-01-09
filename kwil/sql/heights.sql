-- name: SetHeight :exec
UPDATE
    chains
SET
    height = $1
WHERE
    id = $2;

-- name: GetHeight :one
SELECT
    height
FROM
    chains
WHERE
    id = $1;

-- name: GetHeightByName :one
SELECT
    height
FROM
    chains
WHERE
    chain = $1;