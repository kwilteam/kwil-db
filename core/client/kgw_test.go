package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_composeKGWAuthMessage(t *testing.T) {
	param := &gatewayAuthParameter{
		Nonce:          "123456",
		Statement:      "eww",
		IssueAt:        "2023-11-05T22:57:46Z",
		ExpirationTime: "2023-11-05T22:58:16Z",
	}
	msg := composeGatewayAuthMessage(param,
		"https://example.com", "https://example.com/auth", "1", "test-chain")
	want := "https://example.com wants you to sign in with your account:\n\neww\n\nURI: https://example.com/auth\nVersion: 1\nChain ID: test-chain\nNonce: 123456\nIssue At: 2023-11-05T22:57:46Z\nExpiration Time: 2023-11-05T22:58:16Z\n"
	assert.Equal(t, want, msg, "should be equal")
}
