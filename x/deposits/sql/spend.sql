-- name: Spend :exec
UPDATE
   wallets
SET
   balance = balance - $2,
   spent = spent + $2
WHERE
   wallet = $1
   AND balance >= $2;

DO $$ DECLARE found BOOLEAN;

BEGIN IF NOT FOUND THEN RAISE EXCEPTION 'Insufficient balance';

END IF;

END $$;