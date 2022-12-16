-- name: StartWithdrawal :exec
DO $$
DECLARE
    wallet wallet%ROWTYPE := (SELECT * FROM wallets WHERE wallet = $1);
    withdraw_amount NUMERIC(78) := $2;
-- check if there is enough balance
IF wallet.balance - withdraw_amount < 0 THEN
    withdraw_amount := wallet.balance;
    IF (withdraw_amount = 0) AND (wallet.spent = 0) THEN
        RAISE EXCEPTION 'Can not withdraw from an empty wallet';
    END IF;
END IF;

UPDATE wallets
SET balance = balance - withdraw_amount, spent = 0
WHERE wallet_id = wallet.wallet_id;

INSERT INTO withdrawals (wallet_id, amount, fee, expiry, correlation_id)
VALUES (wallet.wallet_id, withdraw_amount, wallet.spent, $3, $4);

END $$;