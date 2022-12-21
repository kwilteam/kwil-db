-- name: GetBalanceAndSpent :one
SELECT
    balance,
    spent,
    id
FROM
    wallets
WHERE
    wallet = $1;

-- name: SetBalanceAndSpent :exec
UPDATE
    wallets
SET
    balance = $2,
    spent = $3
WHERE
    id = $1;

-- name: DecreaseBalance :exec
UPDATE
    wallets
SET
    balance = balance - $2
WHERE
    wallet = $1;