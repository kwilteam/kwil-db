-- name: SetHeight :exec
UPDATE
    chains
SET
    height = $1
WHERE
    code = $2;

-- name: GetHeight :one
SELECT
    height
FROM
    chains
WHERE
    code = $1;

-- name: GetHeightByName :one
SELECT
    height
FROM
    chains
WHERE
    code = $1;