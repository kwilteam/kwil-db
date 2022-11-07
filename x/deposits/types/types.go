package types

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
)

type BlockSubscription interface {
	Unsubscribe()

	Err() <-chan error
}

type Client interface {
	SubscribeBlocks(context.Context, chan<- int64) (BlockSubscription, error)
	GetLatestBlock(context.Context) (int64, error)
	GetContract(string) (Contract, error)
}

type Contract interface {
	GetDeposits(context.Context, int64, int64, string) ([]*Deposit, error)
	GetWithdrawals(context.Context, int64, int64, string) ([]*WithdrawalConfirmation, error)
	ReturnFunds(context.Context, *ecdsa.PrivateKey, string, string, *big.Int, *big.Int) (string, error)
}

type Deposit struct {
	Caller string
	Target string
	Amount string
	Height int64
	Tx     string
	Token  string
}

func (d *Deposit) Serialize() ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 0x00}, b...)
	return b, nil
}

type WithdrawalConfirmation struct {
	Caller   string
	Receiver string
	Amount   string
	Fee      string
	Cid      string
	Height   int64
	Tx       string
	Token    string
}

func (w *WithdrawalConfirmation) Serialize() ([]byte, error) {
	b, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 0x02}, b...)
	return b, nil
}

type WithdrawalRequest struct {
	Wallet     string `json:"wallet"`
	Amount     string `json:"amount"`
	Spent      string `json:"spent"`
	Cid        string `json:"nonce"`
	Expiration int64  `json:"expiration"`
}

type PendingWithdrawal struct {
	Wallet     string `json:"wallet"`
	Amount     string `json:"amount"`
	Fee        string `json:"spent"`
	Cid        string `json:"nonce"`
	Expiration int64  `json:"expiration"`
	Tx         string `json:"tx"`
}

func (wr *WithdrawalRequest) Serialize() ([]byte, error) {
	b, err := json.Marshal(wr)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 0x01}, b...)
	return b, nil
}

type EndBlock struct {
	Height int64 `json:"height"`
}

func (eob *EndBlock) Serialize() ([]byte, error) {
	b, err := json.Marshal(eob)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 0x03}, b...)
	return b, nil
}

type Spend struct {
	Caller string `json:"caller"`
	Amount string `json:"amount"`
}

func (s *Spend) Serialize() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 0x04}, b...)
	return b, nil
}

func Deserialize[T *Deposit | *WithdrawalConfirmation | *WithdrawalRequest | *EndBlock | *Spend](m []byte) (T, error) {
	var t T

	err := json.Unmarshal(m[2:], &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
