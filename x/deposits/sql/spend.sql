
-- name: spend :exec
WITH oldBal AS (
	SELECT balance FROM wallets WHERE wallet = $1
)
SELECT
	CASE
		WHEN oldBal IS NULL THEN 'Wallet ' || $1 || ' does not exist'
		WHEN oldBal - $2 < 0 THEN 'Not enough balance'
		ELSE NULL
	END AS error,
	CASE
		WHEN oldBal IS NOT NULL AND oldBal - $2 >= 0 THEN
			UPDATE wallets
			SET balance = oldBal-$2, spent = (SELECT spent FROM wallets WHERE wallet = $1) + $2
			WHERE wallet = $1
			RETURNING 'Success'
		ELSE
			NULL
	END AS result
FROM oldBal;