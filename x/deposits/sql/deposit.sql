-- name: Deposit :exec
INSERT INTO
    deposits (tx_hash, wallet, amount, height)
VALUES
    ($1, $2, $3, $4);

-- name: CommitDeposits :exec
UPDATE
    wallets
SET
    balance = balance + deposits.amount
FROM
    deposits
WHERE
    deposits.height <= $1
    AND deposits.wallet = wallets.wallet;