-- name: Deposit :exec
INSERT INTO
    deposits (tx_hash, account_address, amount, height)
VALUES
    ($1, $2, $3, $4);

-- name: CommitDeposits :exec
WITH deleted_deposits AS (
    DELETE FROM deposits
    WHERE height <= $1
    RETURNING *
)
INSERT INTO accounts (account_address, balance)
SELECT deleted_deposits.account_address, deleted_deposits.amount
FROM deleted_deposits
ON CONFLICT (account_address) WHERE (account_address is NOT NULL) DO UPDATE SET balance = accounts.balance + (
    SELECT deleted_deposits.amount
    FROM deleted_deposits
    WHERE accounts.account_address = deleted_deposits.account_address
);

-- name: GetDepositIdByTx :one
SELECT id
FROM deposits
WHERE tx_hash = $1;