package gateway

import "github.com/kwilteam/kwil-db/core/crypto/auth"

// GatewayAuth is a request for authentication from the
// kwil gateway.
type GatewayAuth struct {
	Nonce     string          `json:"nonce"`  // identifier for authn session
	Sender    []byte          `json:"sender"` // sender public key
	Signature *auth.Signature `json:"signature"`
}

// GatewayAuthParameter defines the result of GET request for gateway(KGW)
// authentication. It's the parameters that will be used to compose the
// message(SIWE like) to sign.
type GatewayAuthParameter struct {
	Nonce          string `json:"nonce"`
	Statement      string `json:"statement"` // optional
	IssueAt        string `json:"issue_at"`
	ExpirationTime string `json:"expiration_time"`
}
