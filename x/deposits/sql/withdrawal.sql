-- name: NewWithdrawal :exec
INSERT INTO
    withdrawals (correlation_id, wallet_id, amount, fee, expiry)
VALUES
    ($1, $2, $3, $4, $5);

-- name: AddTx :exec
UPDATE
    withdrawals
SET
    tx_hash = $1
WHERE
    correlation_id = $2;