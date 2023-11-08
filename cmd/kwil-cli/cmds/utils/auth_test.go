package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_composeAuthMessage(t *testing.T) {
	param := &authParam{
		Nonce:          "123456",
		Statement:      "eww",
		IssueAt:        "2023-11-05T22:57:46Z",
		ExpirationTime: "2023-11-05T22:58:16Z",
	}
	msg := composeAuthMessage(param,
		"https://example.com", "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7", "https://example.com/auth", "1", "test-chain")
	want := "https://example.com wants you to sign in with your account:\n0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7\n\neww\n\nURI: https://example.com/auth\nVersion: 1\nChain ID: test-chain\nNonce: 123456\nIssue At: 2023-11-05T22:57:46Z\nExpiration Time: 2023-11-05T22:58:16Z\n"
	assert.Equal(t, want, msg, "should be equal")
}
