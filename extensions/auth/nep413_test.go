//go:build auth_nep413 || ext_test

package auth_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/kwilteam/kwil-db/extensions/auth"
	borsch "github.com/near/borsh-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var zeroNonce = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func Test_Nep413(t *testing.T) {
	type testCase struct {
		name        string
		message     string
		nonce       [32]byte
		recipient   string
		callbackUrl string
		signature   string
		hexPubkey   string
		encodeFn    func([]byte) string
		decodeFn    func(string) ([]byte, error)
	}

	testCases := []testCase{
		{
			name:      "meteor wallet",
			message:   "KlNlZSB5b3VyIGlkT1MgcHJvZmlsZSBJRCoKCkRCSUQ6IHg2MjVhODMyYzg0ZjAyZmJlYmIyMjllZTNiNWU2NmI2NzY3ODAyYjI5ZDg3YWNmNzJiOGRkMDVkMQpBY3Rpb246IGdldF93YWxsZXRfaHVtYW5faWQKUGF5bG9hZERpZ2VzdDogZTkxMjJhNjFlMmY4Njg0NmMyM2ViYjc4ZTQ3OGI0ZjNhMjU3NTRjYgoKS3dpbCDwn5aLCg==",
			nonce:     zeroNonce,
			recipient: "idos.network",
			signature: "FtSnvFmYDTYk5nuMo9W0AfPsyIy1Pl4pyttvDWmLBsUH2J1SJU6s1JoJvzjKVf95MRby2kc8+vjvQNLAYRpwCQ==",
			hexPubkey: "bcb7c8d4ae100a39d8d39be9443b96e14dcc3764e682ae9fb004afecc1cba33d", // DhgBrU3N1n36MV9rENSaQgc4xprMgh7N2vY4th8kLRZN
		},
		{
			name:        "mynearwallet",
			message:     "KlNlZSB5b3VyIGlkT1MgcHJvZmlsZSBJRCoKCkRCSUQ6IHg2MjVhODMyYzg0ZjAyZmJlYmIyMjllZTNiNWU2NmI2NzY3ODAyYjI5ZDg3YWNmNzJiOGRkMDVkMQpBY3Rpb246IGdldF93YWxsZXRfaHVtYW5faWQKUGF5bG9hZERpZ2VzdDogZTkxMjJhNjFlMmY4Njg0NmMyM2ViYjc4ZTQ3OGI0ZjNhMjU3NTRjYgoKS3dpbCDwn5aLCg==",
			nonce:       zeroNonce,
			recipient:   "idos.network",
			callbackUrl: `http://localhost:5173/#accountId=juliosantos-staging.testnet&signature=PYogCMrEnbAr7LSVoOYFAz9wZu1IL4Wtj5TL1A%2BJAa05q4RGhtKX8IpghYvFPIkCbcGjeBe%2Fd7INxpfgFaEcDw%3D%3D&publicKey=ed25519%3ADhgBrU3N1n36MV9rENSaQgc4xprMgh7N2vY4th8kLRZN&`,
			signature:   "Jf9lg+2ikw+Xnp6pR74K/kazF+KLPzT5QGb+nualZOGZcDXEcC7cRjsN9iUwdtVDWELJaIh1BYVMHmVYC78iAw==",
			hexPubkey:   "bcb7c8d4ae100a39d8d39be9443b96e14dcc3764e682ae9fb004afecc1cba33d", // DhgBrU3N1n36MV9rENSaQgc4xprMgh7N2vY4th8kLRZN
			//encodeFn:    func(bts []byte) string { return string(bts) },
		},
		{
			name:        "plaintext",
			message:     "*See your idOS profile ID*\n\nDBID: x625a832c84f02fbebb229ee3b5e66b6767802b29d87acf72b8dd05d1\nAction: get_wallet_human_id\nPayloadDigest: e9122a61e2f86846c23ebb78e478b4f3a25754cb\n\nKwil \n",
			nonce:       zeroNonce,
			recipient:   "idos.network",
			callbackUrl: "http://localhost:5173/",
			signature:   "PCvQ1VOrm2uZl2gcE9JPni/3j/C4ZU2kTlFZeofWMCRybT+rLuQ2Zuft3Vv5DKise7/zNqzZIA9noGJQ8RHZBg==",
			hexPubkey:   "bcb7c8d4ae100a39d8d39be9443b96e14dcc3764e682ae9fb004afecc1cba33d", // DhgBrU3N1n36MV9rENSaQgc4xprMgh7N2vY4th8kLRZN
			encodeFn:    strEncode,
			decodeFn:    strDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// converting data
			sig, err := base64.StdEncoding.DecodeString(tc.signature)
			require.NoError(t, err)

			if tc.encodeFn == nil {
				tc.encodeFn = base64.StdEncoding.EncodeToString
			}
			if tc.decodeFn == nil {
				tc.decodeFn = base64.StdEncoding.DecodeString
			}

			pubKey, err := hex.DecodeString(tc.hexPubkey)
			require.NoError(t, err)

			msgBts, err := tc.decodeFn(tc.message)
			require.NoError(t, err)

			serializedPayload, err := borsch.Serialize(auth.Nep413Payload{
				Nonce:       tc.nonce,
				Recipient:   tc.recipient,
				CallbackUrl: &tc.callbackUrl,
			})
			require.NoError(t, err)

			payloadLen := len(serializedPayload)
			// convert to uint16
			payloadLenBytes := []byte{byte(payloadLen >> 8), byte(payloadLen)}

			// test
			err = auth.Nep413Authenticator{
				MsgEncoder: tc.encodeFn,
			}.Verify(pubKey, msgBts, append(payloadLenBytes, append(serializedPayload, sig...)...))
			assert.NoError(t, err, "signature should be valid")
		})
	}
}

func strEncode(bts []byte) string {
	return string(bts)
}

func strDecode(str string) ([]byte, error) {
	return []byte(str), nil
}
