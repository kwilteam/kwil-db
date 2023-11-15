package gatewayclient

import (
	"bytes"
	"fmt"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	types "github.com/kwilteam/kwil-db/core/types/gateway"
)

// kgw handles the authentication with KGW provider.
// KGW is a LB that als provides authentication for Kwil. It only supports HTTP.
// This is not part of core Kwil API, thus we implement it here.
//
// The authentication process is as follows:
// 1. Client starts an authentication session to KGW provider by sending a GET
//    request to /auth endpoint, and KGW will return authn parameters.
// 2. Client composes a message using returned parameters and configuration,
//    then presents the message to the user to sign.
// 3. Then user signs the message and passes the signature back to the client.
// 4. Client identifies itself by sending a POST request to the KGW provider,
//    and KGW will return a cookie if the signature is valid.
// 5. Following requests to the KGW provider should include the cookie for
//    authentication required endpoints.

const (
	kgwAuthVersion    = "1"
	kgwAuthCookieName = "kgw_session"
)

// GatewayAuthSignFunc is the function that signs the authentication message.
type GatewayAuthSignFunc func(message string, signer auth.Signer) (*auth.Signature, error)

// defaultGatewayAuthSignFunc is the default function that signs the message.
// It uses the local signer to sign the message.
func defaultGatewayAuthSignFunc(message string, signer auth.Signer) (*auth.Signature, error) {
	return signer.Sign([]byte(message))
}

// composeGatewayAuthMessage composes the SIWE-like message to sign.
// param is the result of GET request for authentication.
// ALl the other parameters are expected from config.
func composeGatewayAuthMessage(param *types.GatewayAuthParameter, domain string, uri string,
	version string, chainID string) string {
	var msg bytes.Buffer
	msg.WriteString(
		fmt.Sprintf("%s wants you to sign in with your account:\n", domain))
	msg.WriteString("\n")
	if param.Statement != "" {
		msg.WriteString(fmt.Sprintf("%s\n", param.Statement))
	}
	msg.WriteString("\n")
	msg.WriteString(fmt.Sprintf("URI: %s\n", uri))
	msg.WriteString(fmt.Sprintf("Version: %s\n", version))
	msg.WriteString(fmt.Sprintf("Chain ID: %s\n", chainID))
	msg.WriteString(fmt.Sprintf("Nonce: %s\n", param.Nonce))
	msg.WriteString(fmt.Sprintf("Issue At: %s\n", param.IssueAt))
	msg.WriteString(fmt.Sprintf("Expiration Time: %s\n", param.ExpirationTime))
	return msg.String()
}
