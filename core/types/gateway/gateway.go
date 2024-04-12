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
	// client can use those to precheck before signing
	ChainID string `json:"chain_id"` // the chain id of the gateway
	Domain  string `json:"domain"`   // the domain of the gateway
	Version string `json:"version"`  // the authn version used by the gateway
	URI     string `json:"uri"`      // the endpoint used for authn
}
