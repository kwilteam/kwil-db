package deposits

import (
	"context"
	"errors"
	"kwil/x/deposits/types"
	"math/big"
	"math/rand"
)

/*
	The process for Withdrawals is as follows:

	1. The wallet seeking a withdrawal sends a request to the validator.
	2. Function checks things like unique ID, amount requested, etc.
	3. The validator then finds how much the wallet has spent.  The validator will then see how much
	   the wallet has not spent yet.  If the unspent amount is less than the amount requested, the
	   validator will return only the unspent amount, and take the fee.  If the unspent amount is greater
	   than the amount requested, the validator will return the amount requested, and take the fee.
*/

var ErrCantParseAmount = errors.New("can't parse amount")

type WithdrawalResponse struct {
	Tx            string `json:"tx"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	CorrelationId string `json:"correlation_id"`
	Expiration    int64  `json:"expiration"`
	Wallet        string `json:"wallet"`
}

func (d *deposits) Withdraw(ctx context.Context, addr, amt string) (*types.PendingWithdrawal, error) {
	// generate a nonce
	cid := generateCid(10)

	res, err := d.sql.StartWithdrawal(cid, addr, amt, d.we+d.lh)
	if err != nil {
		return nil, err
	}

	// now we need to send the withdrawal request to the blockchain
	pk, err := d.acc.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	// parse amt and fee to *big.Int
	amount, ok := new(big.Int).SetString(res.Amount, 10)
	if !ok {
		return nil, ErrCantParseAmount
	}

	fee, ok := new(big.Int).SetString(res.Fee, 10)
	if !ok {
		return nil, ErrCantParseAmount
	}

	// recip, nonce, amt, fee
	tx, err := d.sc.ReturnFunds(ctx, pk, res.Wallet, res.Cid, amount, fee)
	if err != nil {
		d.log.Errorf("error sending withdrawal request to blockchain: %v", err)
		return nil, err
	}

	// update the withdrawal request with the tx
	// trim off the 0x
	err = d.sql.AddTx(cid, tx[2:])
	if err != nil {
		return nil, err
	}

	return &types.PendingWithdrawal{
		Tx:         tx,
		Amount:     res.Amount,
		Fee:        res.Fee,
		Cid:        res.Cid,
		Expiration: res.Expiration,
		Wallet:     addr,
	}, nil
}

// generateCid generates a correlation id for the withdrawal
func generateCid(l uint8) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, l)
	for i := uint8(0); i < l; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func (d *deposits) GetWithdrawalsForWallet(addr string) ([]*types.PendingWithdrawal, error) {
	return d.sql.GetWithdrawalsForWallet(addr)
}
