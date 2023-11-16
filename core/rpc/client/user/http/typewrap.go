package http

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

// Due to https://protobuf.dev/programming-guides/proto3/#json, protobuf
// fields defined as `int64` will be marshaled as `string` in JSON.
// Those types are wrapped here to unmarshal them correctly.
// NOTE: Should always wrap the protobuf type.

type account txpb.Account

func (a *account) UnmarshalJSON(data []byte) error {
	var acc struct {
		Identifier string `json:"identifier"`
		Balance    string `json:"balance"`
		Nonce      string `json:"nonce"` // int64
	}

	err := json.Unmarshal(data, &acc)
	if err != nil {
		return err
	}

	nonce, err := strconv.ParseInt(acc.Nonce, 10, 64)
	if err != nil {
		return fmt.Errorf("parseNonce: %w", err)
	}

	pk, err := base64.StdEncoding.DecodeString(acc.Identifier)
	if err != nil {
		return fmt.Errorf("parsePublicKey: %w", err)
	}

	a.Balance = acc.Balance
	a.Nonce = nonce
	a.Identifier = pk
	return nil
}

type getAccountResponse struct {
	Account account `json:"account"`
}

type transactionResult txpb.TransactionResult

func (r *transactionResult) UnmarshalJSON(data []byte) error {
	var res struct {
		Code      uint32   `json:"code"`
		Log       string   `json:"log"`
		GasUsed   string   `json:"gas_used"`   // int64
		GasWanted string   `json:"gas_wanted"` // int64
		Data      []byte   `json:"data"`
		Events    [][]byte `json:"events"`
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	r.Code = res.Code
	r.Log = res.Log
	r.GasUsed, err = strconv.ParseInt(res.GasUsed, 10, 64)
	if err != nil {
		return fmt.Errorf("parseGasUsed: %w", err)
	}

	r.GasWanted, err = strconv.ParseInt(res.GasWanted, 10, 64)
	if err != nil {
		return fmt.Errorf("parseGasWanted: %w", err)
	}

	r.Data = res.Data
	r.Events = res.Events
	return nil
}

type transaction txpb.Transaction

func (t *transaction) UnmarshalJSON(data []byte) error {
	var res struct {
		Body struct {
			Payload     []byte `json:"payload"`
			PayloadType string `json:"payload_type"`
			Fee         string `json:"fee"`
			Nonce       string `json:"nonce"`
			ChainID     string `json:"chain_id"`
			Description string `json:"description"`
		} `json:"body"`
		Serialization string `json:"serialization"`
		Signature     struct {
			SignatureBytes []byte `json:"signature_bytes"`
			SignatureType  string `json:"signature_type"`
		} `json:"signature"`
		Sender []byte `json:"sender"`
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	t.Body = &txpb.Transaction_Body{
		Payload:     res.Body.Payload,
		PayloadType: res.Body.PayloadType,
		Fee:         res.Body.Fee,
		ChainId:     res.Body.ChainID,
		Description: res.Body.Description,
	}
	t.Body.Nonce, err = strconv.ParseUint(res.Body.Nonce, 10, 64)
	if err != nil {
		return fmt.Errorf("parseNonce: %w", err)
	}

	t.Serialization = res.Serialization
	t.Signature = &txpb.Signature{
		SignatureBytes: res.Signature.SignatureBytes,
		SignatureType:  res.Signature.SignatureType,
	}
	t.Sender = res.Sender
	return nil
}

type txQueryResponse txpb.TxQueryResponse

func (r *txQueryResponse) UnmarshalJSON(data []byte) error {
	var res struct {
		Hash     []byte            `json:"hash"`
		Height   string            `json:"height"`
		Tx       transaction       `json:"tx"`
		TxResult transactionResult `json:"tx_result"`
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	r.Hash = res.Hash
	r.Height, err = strconv.ParseInt(res.Height, 10, 64)
	if err != nil {
		return fmt.Errorf("parseHeight: %w", err)
	}

	r.Tx = (*txpb.Transaction)(&res.Tx)
	r.TxResult = (*txpb.TransactionResult)(&res.TxResult)
	return nil
}

type validatorJoinsStatusResponse txpb.ValidatorJoinStatusResponse

func (r *validatorJoinsStatusResponse) UnmarshalJSON(data []byte) error {
	var res struct {
		ApprovedValidators [][]byte `json:"approved_validators"`
		PendingValidators  [][]byte `json:"pending_validators"`
		Power              string   `json:"power"`
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	r.ApprovedValidators = res.ApprovedValidators
	r.PendingValidators = res.PendingValidators
	r.Power, err = strconv.ParseInt(res.Power, 10, 64)
	if err != nil {
		return fmt.Errorf("parsePower: %w", err)
	}
	return nil
}

type validator txpb.Validator

func (r *validator) UnmarshalJSON(data []byte) error {
	var res struct {
		PubKey []byte `json:"pubkey"`
		Power  string `json:"power"`
	}

	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	r.Pubkey = res.PubKey
	r.Power, err = strconv.ParseInt(res.Power, 10, 64)
	if err != nil {
		return fmt.Errorf("parsePower: %w", err)
	}
	return nil
}

type currentValidatorsResponse struct {
	Validators []*validator `json:"validators"`
}

type chainInfoResponse txpb.ChainInfoResponse

func (r *chainInfoResponse) UnmarshalJSON(data []byte) error {
	var res struct {
		ChainID string `json:"chain_id"`
		Height  string `json:"height"`
		Hash    string `json:"hash"`
	}
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	r.ChainId = res.ChainID
	r.Height, err = strconv.ParseUint(res.Height, 10, 64)
	if err != nil {
		return fmt.Errorf("parseHeight: %w", err)
	}
	r.Hash = res.Hash
	return nil
}
