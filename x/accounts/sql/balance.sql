-- name: DecreaseBalance :exec
UPDATE
    accounts
SET
    balance = balance - $2
WHERE
    account_address = $1;

-- name: IncreaseBalance :exec
UPDATE
    accounts
SET
    balance = balance + $2
WHERE
    account_address = $1;