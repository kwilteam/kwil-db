-- name: GetAccount :one
SELECT
    balance,
    spent,
    id,
    nonce
FROM
    accounts
WHERE
    account_address = $1;

-- name: UpdateAccountById :exec
UPDATE
    accounts
SET
    balance = $1,
    spent = $2,
    nonce = $3
WHERE
    id = $4;

-- name: UpdateAccountByAddress :exec
UPDATE
    accounts
SET
    balance = $1,
    spent = $2,
    nonce = $3
WHERE
    account_address = $4;