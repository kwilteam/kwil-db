-- name: Spend :exec
WITH previous_nonce AS (
   SELECT
      nonce
   FROM
      accounts
   WHERE
      account_address = $1
   FOR UPDATE
)
UPDATE
   accounts a
SET
   balance = a.balance - $2,
   spent = a.spent + $2,
   nonce = $3
WHERE
   a.account_address = $1
   AND a.balance >= $2;

DO $$;

BEGIN
   IF NOT FOUND THEN RAISE
   EXCEPTION
      'Insufficient balance';
   IF previous_nonce + 1 <> $3 THEN RAISE
   EXCEPTION
      'Invalid nonce';

END IF;

END $$;