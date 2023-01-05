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