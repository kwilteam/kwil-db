package types

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"kwil/crypto"
	"kwil/crypto/auth"
)

type TestPayload struct {
	Key   string
	Value string
}

func (p *TestPayload) MarshalBinary() ([]byte, error) {
	return []byte(fmt.Sprintf("%s=%s", p.Key, p.Value)), nil
}

func (p *TestPayload) UnmarshalBinary(data []byte) error {
	key, value, ok := bytes.Cut(data, []byte("="))
	if !ok {
		return errors.New("invalid payload format")
	}
	p.Key = string(key)
	p.Value = string(value)
	return nil
}

func (p *TestPayload) Type() PayloadType {
	return "test"
}

func TestTransactionSerialization(t *testing.T) {
	t.Parallel()
	payload := &TestPayload{
		Key:   "dummy",
		Value: "data",
	}

	payloadBts, err := payload.MarshalBinary()
	require.NoError(t, err)

	defDesc := "You are signing a kwil transaction of type test"
	longDesc := "dfhjdshfksdhfkshdkfjsdbfhjsbdfhsbdkfhsbdfhgjsdbfhjsdbjhfbsdkfhjsdfdsbfjhsdbfhjgdsfvbhgsdbfjhsdbfhsjdbfhsdbfshdjbfhdsfvbhsdbfhsdbfshdgfvhsdbfhsdbfhdbfhasfhwegsfwegsfhwedfegsfysegdfysegfhesgfyuwegfywesgfyswegfywuegfuyse"

	chainID := "test-chain"
	nonce := uint64(1)

	// Tests for SerializeMsg:
	// 1. Invalid serialization type
	// 2. Default serialization type
	// 3. Long description

	testcase := []struct {
		name        string
		serType     SignedMsgSerializationType
		desc        string
		expectError bool
	}{
		{
			name:        "invalid serialization type",
			serType:     "invalid",
			desc:        defDesc,
			expectError: true,
		},
		{
			name:        "default serialization type",
			serType:     DefaultSignedMsgSerType, // concat
			desc:        defDesc,
			expectError: false,
		},
		{
			name:        "long description",
			serType:     DefaultSignedMsgSerType, // concat
			desc:        longDesc,
			expectError: true,
		},
	}

	for _, tt := range testcase {
		txBody := &TransactionBody{
			Description: tt.desc,
			Payload:     payloadBts,
			PayloadType: payload.Type(),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
			ChainID:     chainID,
		}

		msg, err := txBody.SerializeMsg(tt.serType)
		if tt.expectError {
			require.Error(t, err, tt.name)
			continue
		}

		require.NoError(t, err, tt.name)
		require.NotEmpty(t, msg, tt.name)
		fmt.Println(tt.name, ":", string(msg))
	}

}

func TestTransactionSign(t *testing.T) {
	t.Parallel()
	// Test scenarios:
	// 1. Invalid signer
	// 2. Valid signer: Sign + Verify
	// Signers to test: EthPersonalSigner, Ed25519Signer
	// 3. Invalid serialization type

	payload := &TestPayload{
		Key:   "dummy",
		Value: "data",
	}

	payloadBts, err := payload.MarshalBinary()
	require.NoError(t, err)

	testcases := []struct {
		name          string
		serType       SignedMsgSerializationType
		signer        auth.Signer
		authenticator auth.Authenticator
		expectError   bool
	}{
		{
			name:          "valid ed25519 signer",
			serType:       DefaultSignedMsgSerType,
			signer:        ed25519Signer(t),
			authenticator: auth.Ed25519Authenticator{},
			expectError:   false,
		},
		{
			name:          "valid eth personal signer",
			serType:       DefaultSignedMsgSerType,
			signer:        secp256k1Signer(t),
			authenticator: auth.EthSecp256k1Authenticator{},
			expectError:   false,
		},
		{
			name:          "invalid serialization type",
			serType:       "invalid",
			signer:        secp256k1Signer(t),
			authenticator: auth.EthSecp256k1Authenticator{},
			expectError:   true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {

			tx := &Transaction{
				Body: &TransactionBody{
					Payload:     payloadBts,
					PayloadType: payload.Type(),
					Fee:         big.NewInt(0),
					Nonce:       1,
					ChainID:     "test-chain",
				},
				Serialization: tt.serType,
			}

			err = tx.Sign(tt.signer)
			if tt.expectError {
				require.Error(t, err, tt.name)
				return
			}

			require.NoError(t, err, tt.name)

			// Verify the signature
			msg, err := tx.SerializeMsg()
			require.NoError(t, err, "failed to serialize transaction")

			err = tt.authenticator.Verify(tx.Sender, msg, tx.Signature.Data)
			require.NoError(t, err, "signature verification failed")
		})
	}
}

func TestTransactionBodyMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	payload := &TestPayload{
		Key:   "dummy",
		Value: "data",
	}

	payloadBts, err := payload.MarshalBinary()
	require.NoError(t, err)

	testcases := []struct {
		name string
		body *TransactionBody
	}{
		{
			name: "valid transaction body",
			body: &TransactionBody{
				Description: "",
				Payload:     payloadBts,
				PayloadType: payload.Type(),
				Fee:         big.NewInt(1000000),
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
		{
			name: "empty payload",
			body: &TransactionBody{
				Description: "You are signing a kwil transaction of type test",
				Payload:     nil,
				PayloadType: payload.Type(),
				Fee:         big.NewInt(1000000),
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
		{
			name: "empty description",
			body: &TransactionBody{
				Description: "",
				Payload:     payloadBts,
				PayloadType: payload.Type(),
				Fee:         big.NewInt(0),
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
		{
			name: "empty Fee",
			body: &TransactionBody{
				Description: "You are signing a kwil transaction of type test",
				Payload:     payloadBts,
				PayloadType: payload.Type(),
				Fee:         nil,
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
		{
			name: "zero fee",
			body: &TransactionBody{
				Description: "You are signing a kwil transaction of type test",
				Payload:     payloadBts,
				PayloadType: payload.Type(),
				Fee:         big.NewInt(0),
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
		{
			name: "empty payload type",
			body: &TransactionBody{
				Description: "You are signing a kwil transaction of type test",
				Payload:     payloadBts,
				PayloadType: "",
				Fee:         big.NewInt(1000000),
				Nonce:       1,
				ChainID:     "test-chain",
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.body.MarshalBinary()
			require.NoError(t, err)

			newBody := &TransactionBody{}
			err = newBody.UnmarshalBinary(data)
			require.NoError(t, err)

			require.Equal(t, tt.body.Description, newBody.Description)
			require.Equal(t, tt.body.PayloadType, newBody.PayloadType)
			fee := tt.body.Fee
			if fee == nil {
				fee = big.NewInt(0)
			}
			require.Equal(t, fee, newBody.Fee)
			require.Equal(t, tt.body.Nonce, newBody.Nonce)
			require.Equal(t, tt.body.ChainID, newBody.ChainID)
		})
	}
}

func TestTransactionMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	payload := &TestPayload{
		Key:   "dummy",
		Value: "data",
	}

	payloadBts, err := payload.MarshalBinary()
	require.NoError(t, err)

	signer := secp256k1Signer(t)
	sender := signer.Identity()

	require.NoError(t, err)

	testcases := []struct {
		name        string
		signer      auth.Signer
		expectError bool
		fn          func(t *testing.T) *Transaction
	}{
		{
			name:        "valid transaction",
			signer:      secp256k1Signer(t),
			expectError: false,
			fn: func(t *testing.T) *Transaction {
				// sign tx
				tx := &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Serialization: DefaultSignedMsgSerType,
				}

				err := tx.Sign(secp256k1Signer(t))
				require.NoError(t, err)

				return tx
			},
		},
		{
			name: "empty signature",
			fn: func(t *testing.T) *Transaction {
				tx := &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Serialization: DefaultSignedMsgSerType,
				}

				err := tx.Sign(secp256k1Signer(t))
				require.NoError(t, err)

				return tx
			},
			expectError: false,
		},
		{
			name:   "empty signature type",
			signer: secp256k1Signer(t),
			fn: func(t *testing.T) *Transaction {
				tx := &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Serialization: DefaultSignedMsgSerType,
					Signature: &auth.Signature{
						Data: []byte("signature"),
					},
					Sender: sender,
				}

				return tx
			},
			expectError: false,
		},
		{
			name: "empty signature data",
			fn: func(t *testing.T) *Transaction {
				tx := &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Serialization: DefaultSignedMsgSerType,
					Signature: &auth.Signature{
						Type: "secp256k1",
					},
					Sender: sender,
				}

				return tx
			},
			expectError: false,
		},
		{
			name: "empty body",
			fn: func(t *testing.T) *Transaction {
				return &Transaction{
					Body: nil,
					Signature: &auth.Signature{
						Data: []byte("signature"),
						Type: "secp256k1",
					},
					Sender:        sender,
					Serialization: DefaultSignedMsgSerType,
				}
			},
		},
		{
			name: "empty sender",
			fn: func(t *testing.T) *Transaction {
				return &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Signature: &auth.Signature{
						Data: []byte("signature"),
						Type: "secp256k1",
					},
					Sender:        nil,
					Serialization: DefaultSignedMsgSerType,
				}
			},
		},
		{
			name: "empty serialization type",
			fn: func(t *testing.T) *Transaction {
				return &Transaction{
					Body: &TransactionBody{
						Description: "You are signing a kwil transaction of type test",
						Payload:     payloadBts,
						PayloadType: payload.Type(),
						Fee:         big.NewInt(19990),
						Nonce:       1,
						ChainID:     "test-chain",
					},
					Signature: &auth.Signature{
						Data: []byte("signature"),
						Type: "secp256k1",
					},
					Sender: sender,
				}
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.fn(t)

			data, err := tx.MarshalBinary()
			require.NoError(t, err)

			newTx := &Transaction{}
			err = newTx.UnmarshalBinary(data)
			require.NoError(t, err)

			require.Equal(t, tx, newTx)
		})
	}
}

func secp256k1Signer(t *testing.T) *auth.EthPersonalSigner {
	privKey, _, err := crypto.GenerateSecp256k1Key(nil)
	require.NoError(t, err)

	privKeyBytes := privKey.Bytes()
	k, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	require.NoError(t, err)

	return &auth.EthPersonalSigner{Key: *k}
}

func ed25519Signer(t *testing.T) *auth.Ed25519Signer {
	privKey, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)

	pBytes := privKey.Bytes()
	k, err := crypto.UnmarshalEd25519PrivateKey(pBytes)
	require.NoError(t, err)

	return &auth.Ed25519Signer{Ed25519PrivateKey: *k}
}
