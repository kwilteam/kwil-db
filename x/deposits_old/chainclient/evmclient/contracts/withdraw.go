package contracts

import (
	"context"
	"crypto/ecdsa"
	"kwil/abi"
	ct "kwil/x/deposits_old/types"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// RetrunFunds calls the returnDeposit function
func (c *contract) ReturnFunds(ctx context.Context, key *ecdsa.PrivateKey, recipient, nonce string, amt *big.Int, fee *big.Int) (string, error) {

	txOpts, err := bind.NewKeyedTransactorWithChainID(key, c.cid)
	if err != nil {
		return "", err
	}

	res, err := c.ctr.ReturnDeposit(txOpts, common.HexToAddress(recipient), amt, fee, nonce)
	if err != nil {
		return "", err
	}

	return res.Hash().String(), nil
}

func (c *contract) GetWithdrawals(ctx context.Context, from, to int64, addr string) ([]*ct.WithdrawalConfirmation, error) {
	end := uint64(to)
	queryOpts := &bind.FilterOpts{Context: ctx, Start: uint64(from), End: &end}

	ads := common.HexToAddress(addr)

	edi, err := c.ctr.FilterWithdrawal(queryOpts, []common.Address{ads})
	if err != nil {
		return nil, err
	}

	return convertWithdrawals(edi, c.token), nil
}

func convertWithdrawals(edi *abi.EscrowWithdrawalIterator, token string) []*ct.WithdrawalConfirmation {
	var withdrawals []*ct.WithdrawalConfirmation
	for {

		if !edi.Next() {
			break
		} else {
			withdrawals = append(withdrawals, escToWithdrawal(edi.Event, token))
		}
	}

	return withdrawals
}

func escToWithdrawal(ed *abi.EscrowWithdrawal, token string) *ct.WithdrawalConfirmation {

	return &ct.WithdrawalConfirmation{
		Amount:   ed.Amount.String(),
		Caller:   ed.Caller.Hex(),
		Height:   int64(ed.Raw.BlockNumber),
		Receiver: ed.Receiver.Hex(),
		Tx:       ed.Raw.TxHash.Hex(),
		Token:    token,
		Fee:      ed.Fee.String(),
		Cid:      ed.Nonce,
	}
}
