CREATE TABLE IF NOT EXISTS wallets (
		wallet_id SERIAL PRIMARY KEY,
		wallet VARCHAR(44) NOT NULL UNIQUE,
		balance NUMERIC(78) DEFAULT '0' NOT NULL,
		spent NUMERIC(78) DEFAULT '0' NOT NULL
	);

	CREATE TABLE IF NOT EXISTS deposits (
		deposit_id SERIAL PRIMARY KEY,
		txid VARCHAR(64) NOT NULL UNIQUE,
		wallet VARCHAR(44) NOT NULL,
		amount NUMERIC(78),
		height BIGINT
	);

	CREATE INDEX IF NOT EXISTS deposit_height ON deposits(height);

	CREATE TABLE IF NOT EXISTS withdrawals (
		withdrawal_id SERIAL PRIMARY KEY,
		correlation_id VARCHAR(10) NOT NULL UNIQUE,
		wallet_id INTEGER,
		amount NUMERIC(78),
		fee NUMERIC(78),
		expiry BIGINT,
		tx VARCHAR(64),
		FOREIGN KEY(wallet_id) REFERENCES wallets(wallet_id)
	);

	CREATE INDEX IF NOT EXISTS expiration ON withdrawals(expiry);

	-- the height table is meant to be a key value store for the current height
	CREATE TABLE IF NOT EXISTS height (
		height BIGINT PRIMARY KEY
	);

	CREATE OR REPLACE FUNCTION set_height(h BIGINT) RETURNS VOID AS $$
	BEGIN
		IF EXISTS (SELECT 1 FROM height) THEN
			UPDATE height SET height = h;
		ELSE
			INSERT INTO height VALUES (h);
		END IF;
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION get_height() RETURNS BIGINT AS $$
	BEGIN
		-- get the height from the height table	
		RETURN (SELECT * FROM height);
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION deposit(t VARCHAR(64), w VARCHAR(44), a NUMERIC(78), h BIGINT) RETURNS VOID AS $$
	BEGIN
		INSERT INTO deposits (txid, wallet, amount, height) VALUES (t, w, a, h);
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION spend_money(addr VARCHAR(44), amt numeric(78)) RETURNS void AS $$
		DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr); 
		BEGIN
			-- check old balance exists
			IF oldBal IS NULL THEN
				RAISE EXCEPTION 'Wallet % does not exist', addr;
			END IF;

			-- TODO: Add sequence ID
			IF oldBal - amt < 0 THEN RAISE EXCEPTION 'not enough balance'; END IF;
			UPDATE wallets SET balance = oldBal-amt, spent = (SELECT spent FROM wallets WHERE wallet = addr) + amt WHERE wallet = addr;
		
		END; 
		$$ LANGUAGE plpgsql;
	
	CREATE OR REPLACE FUNCTION commit_deposits(ht BIGINT) RETURNS void AS $$
		DECLARE d deposits%ROWTYPE;
		BEGIN
			-- loop through all deposits at or below the height.  If the wallet already has a balance, it will add the deposit to the balance
			-- If the wallet does not have a balance, it will create a new balance.  Then it will delete the deposit
			FOR d IN SELECT * FROM deposits WHERE height <= ht LOOP
				INSERT INTO wallets (wallet, balance) VALUES (d.wallet, d.amount) ON CONFLICT (wallet) DO UPDATE SET balance = wallets.balance + d.amount;
				DELETE FROM deposits WHERE deposit_id = d.deposit_id;
			END LOOP;

		END;
		$$ LANGUAGE plpgsql;
	
	CREATE OR REPLACE FUNCTION start_withdrawal(addr varchar(44), cid varchar(10), amt numeric(78), expiry BIGINT) RETURNS table (ret_addr varchar(44), ret_cid varchar(10), ret_amount numeric(78), ret_fee numeric(78), ret_expiration BiGINT) AS $$
		DECLARE wid integer = (SELECT wallet_id FROM wallets WHERE wallet = addr);
		DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr);
		DECLARE spent NUMERIC(78) = (SELECT spent FROM wallets WHERE wallet = addr);

		BEGIN
			-- If oldBal > amt, then make amt = oldBal and set oldBal to 0
			IF oldBal - amt < 0 THEN amt = oldBal; oldBal = 0; ELSE oldBal = oldBal - amt; END IF;
			-- If both spent and amt are 0, then raise an exception
			IF spent + amt = 0 THEN RAISE EXCEPTION 'cannot withdraw with 0 balance and spent'; END IF;

			-- Now we insert into withdrawals
			INSERT INTO withdrawals (correlation_id, wallet_id, amount, fee, expiry) VALUES (cid, wid, amt, spent, expiry);
			-- Now we update the wallets table
			UPDATE wallets SET balance = oldBal, spent = 0 WHERE wallet_id = wid;

			RETURN QUERY SELECT addr, cid, amt, spent, expiry;
		END;

		$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION add_tx(cid VARCHAR(10), txh VARCHAR(64)) RETURNS void AS $$ 
		BEGIN
			UPDATE withdrawals SET tx = txh WHERE correlation_id = cid;
		END;
		$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION remove_balance(addr varchar(44), amount numeric(78)) RETURNS void AS $$
		-- subtract the balance from the wallet
		DECLARE oldBal NUMERIC(78) = (SELECT balance FROM wallets WHERE wallet = addr);

		BEGIN
			IF oldBal - amount < 0 THEN RAISE EXCEPTION 'not enough balance'; END IF;
			UPDATE wallets SET balance = oldBal - amount WHERE wallet = addr;
		END;
		$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION expire(height BIGINT) RETURNS VOID AS $$
		DECLARE w withdrawals%ROWTYPE;
		BEGIN
			-- loop through all withdrawals at or below the height.  Transfer the amount back to the balance, and the fee back to the spent, and delete the withdrawal
	
			FOR w IN SELECT * FROM withdrawals WHERE expiry <= height LOOP
				UPDATE wallets SET balance = balance + w.amount, spent = spent + w.fee WHERE wallet_id = w.wallet_id;
				DELETE FROM withdrawals WHERE withdrawal_id = w.withdrawal_id;
			END LOOP;
	
			END;
		$$ LANGUAGE plpgsql;
	
	-- finish_withdrawal will rewturn a boolean for whether or not there was a withdrawal with that correlation_id
	CREATE OR REPLACE FUNCTION finish_withdrawal(n VARCHAR(10)) RETURNS BOOLEAN AS $$
	BEGIN
		-- delete the withdrawal with the given cid
		DELETE FROM withdrawals WHERE correlation_id = n;
		-- return true if there was a withdrawal with that cid
		RETURN FOUND;
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION commit_height(ht BIGINT) RETURNS VOID AS $$
	BEGIN
		PERFORM commit_deposits(ht);
		PERFORM expire(ht);
		PERFORM set_height(ht+1);
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION get_balance(addr VARCHAR(44)) RETURNS NUMERIC(78) AS $$
	BEGIN
		RETURN (SELECT balance FROM wallets WHERE wallet = addr);
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION get_spent(addr VARCHAR(44)) RETURNS NUMERIC(78) AS $$
	BEGIN
		RETURN (SELECT spent FROM wallets WHERE wallet = addr);
	END;	
	$$ LANGUAGE plpgsql;

	-- get_balance_and_spent will return the balance and spent for a given wallet from the wallets table
	CREATE OR REPLACE FUNCTION get_balance_and_spent(addr VARCHAR(44)) RETURNS TABLE (bal NUMERIC(78), sp NUMERIC(78)) AS $$
	BEGIN
		RETURN QUERY SELECT balance, spent FROM wallets WHERE wallet = addr;
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION get_withdrawals_addr(addr VARCHAR(44)) RETURNS TABLE (n VARCHAR(10), a NUMERIC(78), s NUMERIC(78), e BIGINT, t varchar(64)) AS $$
	BEGIN
		RETURN QUERY SELECT correlation_id, amount, fee, expiry, tx FROM withdrawals WHERE wallet_id = (SELECT wallet_id FROM wallets WHERE wallet = addr);
	END;
	$$ LANGUAGE plpgsql;
	
	CREATE OR REPLACE FUNCTION get_all_withdrawals(h BIGINT) RETURNS TABLE (n VARCHAR(10), a NUMERIC(78), s NUMERIC(78), e BIGINT, w varchar(44)) AS $$
	-- must return cid, amount, fee, expiry, and wallet (based on wallet_id)
	BEGIN
		-- Get withdrawals at or before height h and join with wallets to get the wallet
		-- Return the cid, amount, fee, expiry, and wallet
		RETURN QUERY SELECT correlation_id, amount, fee, expiry, wallet FROM withdrawals JOIN wallets ON withdrawals.wallet_id = wallets.wallet_id WHERE expiry <= h;
	END;
	$$ LANGUAGE plpgsql;