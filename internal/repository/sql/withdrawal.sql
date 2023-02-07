-- name: NewWithdrawal :exec
INSERT INTO
    withdrawals (correlation_id, account_id, amount, fee, expiry)
VALUES
    ($1, (SELECT id FROM accounts WHERE account_address = $2), $3, $4, $5);

-- name: AddTxHashToWithdrawal :exec
UPDATE
    withdrawals
SET
    tx_hash = $1
WHERE
    correlation_id = $2;

-- name: ExpireWithdrawals :exec
UPDATE
    accounts a
SET
    balance = balance + wd.amount,
    spent = spent + wd.fee
FROM
    withdrawals wd
WHERE
    wd.account_id = a.id
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