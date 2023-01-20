-- name: Deposit :exec
INSERT INTO
    deposits (tx_hash, account_address, amount, height)
VALUES
    ($1, $2, $3, $4);

-- name: CommitDeposits :exec
WITH deleted_deposits AS (
    SELECT account_address, SUM(amount) as total_amount
    FROM deposits
    WHERE height <= $1
    GROUP BY account_address
)
INSERT INTO accounts (account_address, balance)
SELECT deleted_deposits.account_address, deleted_deposits.total_amount
FROM deleted_deposits
ON CONFLICT (account_address) WHERE (account_address is NOT NULL) DO UPDATE
SET balance = accounts.balance + (
    SELECT SUM(deleted_deposits.total_amount)
    FROM deleted_deposits
    WHERE accounts.account_address = deleted_deposits.account_address
);

-- name: GetDepositIdByTx :one
SELECT id
FROM deposits
WHERE tx_hash = $1;
