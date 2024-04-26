package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

type transaction struct {
	Raw         []byte
	Tx          *transactions.Transaction
	WithPayload bool
}

// remarshalPayload attempt to decode and remarshal the payload from RLP to JSON.
func (t *transaction) remarshalPayload() (json.RawMessage, bool) {
	payloadObject, err := transactions.UnmarshalPayload(t.Tx.Body.PayloadType, t.Tx.Body.Payload)
	if err != nil {
		return nil, false
	}
	payloadJSON, err := json.Marshal(payloadObject)
	if err != nil {
		return nil, false
	}
	return payloadJSON, true
}

func (t *transaction) MarshalJSON() ([]byte, error) {
	if t.WithPayload {
		payloadJSON, _ := t.remarshalPayload()
		tx := struct {
			Tx      *transactions.Transaction `json:"tx"`
			Payload json.RawMessage           `json:"payload_decoded"`
		}{
			Tx:      t.Tx,
			Payload: payloadJSON,
		}
		return json.MarshalIndent(tx, "", "  ")
	}

	// Decode a fresh Transaction instance and zero out the Payload.
	tx, err := decodeTx(t.Raw)
	if err != nil {
		return nil, err
	}
	tx.Tx.Body.Payload = nil
	return json.MarshalIndent(tx, "", "  ")
}

func (t *transaction) MarshalText() ([]byte, error) {
	txHash := sha256.Sum256(t.Raw) // tmhash is sha256
	msg := fmt.Sprintf(`Transaction ID: %x
Sender: %s
Description: %s
Payload type: %s
ChainID: %v
Fee: %s
Nonce: %d
Signature type: %s
Signature: %s
`,
		txHash,
		hex.EncodeToString(t.Tx.Sender), // hex because it's an address or pubkey, probably address
		t.Tx.Body.Description,
		t.Tx.Body.PayloadType,
		t.Tx.Body.ChainID,
		t.Tx.Body.Fee,
		t.Tx.Body.Nonce,
		t.Tx.Signature.Type,
		base64.StdEncoding.EncodeToString(t.Tx.Signature.Signature),
	)

	if t.WithPayload { // put it at the end regardless since it' can be big
		// First try to decode the transaction (RLP), then create readable JSON
		// for its display. If either fails, show it as base64.
		payloadJSON, ok := t.remarshalPayload()
		if !ok {
			msg += "Payload (invalid): " + base64.StdEncoding.EncodeToString(t.Tx.Body.Payload) + "\n"
		} else {
			msg += "Payload (json): " + string(payloadJSON) + "\n"
		}
	}

	return []byte(msg), nil
}

type respChainInfo struct {
	Info *types.ChainInfo
}

func (r *respChainInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Info)
}

func (r *respChainInfo) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf(`Chain ID: %s
Height: %d
Hash: %s
`,
		r.Info.ChainID,
		r.Info.BlockHeight,
		r.Info.BlockHash,
	)

	return []byte(msg), nil
}

// respKwilCliConfig is used to represent a kwil-cli config in cli
type respKwilCliConfig struct {
	cfg *config.KwilCliConfig
}

func (r *respKwilCliConfig) MarshalJSON() ([]byte, error) {
	cfg := r.cfg.ToPersistedConfig()
	cfg.PrivateKey = "***"
	return json.Marshal(cfg)
}

func (r *respKwilCliConfig) MarshalText() ([]byte, error) {
	var msg bytes.Buffer
	cfg := r.cfg.ToPersistedConfig()
	cfg.PrivateKey = "***"

	v := reflect.ValueOf(cfg)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		msg.WriteString(fmt.Sprintf("%s: %v\n", field.Name, fieldValue))
	}

	return msg.Bytes(), nil
}
