package types

import (
	"fmt"
	"kwil/crypto"
	"kwil/crypto/auth"
	"kwil/node/types/serialize"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestPayload struct {
	Key   string
	Value string
}

func (p *TestPayload) MarshalBinary() ([]byte, error) {
	return serialize.Encode(p)
}

func (p *TestPayload) UnmarshalBinary(data []byte) error {
	return serialize.Decode(data, p)
}

func (p *TestPayload) Type() PayloadType {
	return "test"
}

func TestTransactionSerialization(t *testing.T) {
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
	return &auth.Ed25519Signer{Ed25519PrivateKey: *k}
}
