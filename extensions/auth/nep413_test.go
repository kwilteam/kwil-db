//go:build auth_nep413 || ext_test

package auth_test

import (
	"encoding/base64"
	"testing"

	"github.com/kwilteam/kwil-db/extensions/auth"
	borsch "github.com/near/borsh-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Nep413(t *testing.T) {
	// initial data
	message := "idOS authentication"
	payload := auth.Nep413Payload{
		Recipient: "idos.network",
		Nonce:     [32]byte{5, 233, 107, 175, 203, 182, 15, 111, 97, 146, 18, 10, 118, 80, 180, 9, 186, 39, 255, 93, 36, 218, 196, 25, 72, 177, 237, 28, 173, 75, 17, 31},
	}
	b64Sig := "Ni+rXvOtyzRr7X+qtvQ9+iJUu2e8L/e6cPjSzOYr+6W22chVnptTW0QqTUhFgKUbgPwd2tTcfB1D9Q+0Xb+sBg=="
	pubkey := []byte{0x6c, 0x4f, 0x1b, 0xe1, 0xc1, 0xad, 0x86, 0xfc, 0xff, 0x83,
		0x90, 0x9b, 0xf9, 0x5c, 0x68, 0xb8, 0xe9, 0xe3, 0xc7, 0x5f, 0x52, 0x57,
		0x3, 0xf5, 0x3e, 0x9f, 0x27, 0x51, 0x84, 0xbb, 0x56, 0x57} // 8HnzkUaX21h99idPghFajoV3JZvy3SmJ4mqVwSVfLByg

	// converting data
	sig, err := base64.StdEncoding.DecodeString(b64Sig)
	require.NoError(t, err)

	serializedPayload, err := borsch.Serialize(payload)
	require.NoError(t, err)

	payloadLen := len(serializedPayload)
	// convert to uint16
	payloadLenBytes := []byte{byte(payloadLen >> 8), byte(payloadLen)}

	// test
	err = auth.Nep413Authenticator{
		MsgEncoder: testMsgEncoder,
	}.Verify(pubkey, []byte(message), append(payloadLenBytes, append(serializedPayload, sig...)...))
	assert.NoError(t, err, "signature should be valid")

	addr, err := auth.Nep413Authenticator{}.Address(pubkey)
	assert.NoError(t, err, "address should be valid")
	assert.Equal(t, "6c4f1be1c1ad86fcff83909bf95c68b8e9e3c75f525703f53e9f275184bb5657", addr, "address should be valid")
}

// testMsgEncoder is a function that encodes a message into a string.
// Since the unit test tests a plaintext message (instead of base64).
// This should probably be replaced with base64 encoding once someone on
// the team is able to generate signatures via MyNEARWallet.
func testMsgEncoder(msg []byte) string {
	return string(msg)
}
