-- name: Deposit :exec
INSERT INTO
    deposits (tx_hash, wallet, amount, height)
VALUES
    ($1, $2, $3, $4);

-- name: CommitDeposits :exec
WITH deleted_deposits AS (
    DELETE FROM deposits
    WHERE height <= $1
    RETURNING *
)
INSERT INTO wallets (wallet, balance)
SELECT deleted_deposits.wallet, deleted_deposits.amount
FROM deleted_deposits
ON CONFLICT (wallet) DO UPDATE SET balance = wallets.balance + (
    SELECT deleted_deposits.amount
    FROM deleted_deposits
    WHERE wallets.wallet = deleted_deposits.wallet
);


-- name: GetDepositByTx :one
SELECT id
FROM deposits
WHERE tx_hash = $1;