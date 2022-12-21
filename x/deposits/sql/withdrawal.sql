-- name: NewWithdrawal :exec
INSERT INTO
    withdrawals (correlation_id, wallet_id, amount, fee, expiry)
VALUES
    ($1, $2, $3, $4, $5);

-- name: AddTxHash :exec
UPDATE
    withdrawals
SET
    tx_hash = $1
WHERE
    correlation_id = $2;

-- name: ExpireWithdrawals :exec
UPDATE
    wallets w
SET
    balance = balance + wd.amount,
    spent = spent + wd.fee
FROM
    withdrawals wd
WHERE
    wd.wallet_id = w.id
    AND wd.expiry <= $1;

DELETE FROM
    withdrawals
WHERE
    expiry <= $1;

-- name: ConfirmWithdrawal :exec
DELETE FROM
    withdrawals
WHERE
    correlation_id = $1;

DO $$;

BEGIN
    IF NOT FOUND THEN RAISE
    EXCEPTION
        'Correlation ID not found';

END IF;

END $$;