package types

import (
	"encoding/json"
)

type WithdrawalRequest struct {
	Wallet     string `json:"wallet"`
	Amount     string `json:"amount"`
	Spent      string `json:"spent"`
	Nonce      string `json:"nonce"`
	Expiration int64  `json:"expiration"`
}

func (wr *WithdrawalRequest) Serialize() ([]byte, error) {
	b, err := json.Marshal(wr)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 1}, b...)
	return b, nil
}

type EndOfBlock struct {
	Height int64 `json:"height"`
}

func (eob *EndOfBlock) Serialize() ([]byte, error) {
	b, err := json.Marshal(eob)
	if err != nil {
		return nil, err
	}

	// add magic byte and type
	b = append([]byte{0, 3}, b...)
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
	b = append([]byte{0, 4}, b...)
	return b, nil
}

func Deserialize[T *Deposit | *WithdrawalConfirmation | *WithdrawalRequest | *EndOfBlock | *Spend](m []byte) (T, error) {
	var t T

	err := json.Unmarshal(m[2:], &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
