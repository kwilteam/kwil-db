package chain

import (
	"encoding/json"
)

// Original message sent by secondary node requesting authentication
type AuthRequest struct {
	Address string `json:"address"`
}

func (ar *AuthRequest) Bytes() ([]byte, error) {
	// Convert ar to bytes
	return json.Marshal(ar)
}

// AuthChallenge is used by the primary node to challenge the secondary node
type AuthChallenge struct {
	Valid bool   `json:"valid"`
	Nonce string `json:"nonce"`
}

func (ac *AuthChallenge) Bytes() ([]byte, error) {
	// Convert ar to bytes
	return json.Marshal(ac)
}

// AuthChallengeResponse is used by the secondary node to respond to the primary node
type AuthChallengeResponse struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

func (acr *AuthChallengeResponse) Bytes() ([]byte, error) {
	return json.Marshal(acr)
}

type AuthResponse struct {
	Nonce string `json:"nonce"`
	Valid bool   `json:"valid"`
}

func (ar *AuthResponse) Bytes() ([]byte, error) {
	return json.Marshal(ar)
}
