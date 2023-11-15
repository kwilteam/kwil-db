package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
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
	kgwAuthEndpoint   = "/auth"
	kgwAuthCookieName = "kgw_session"
)

// gatewayAuthParameter defines the result of GET request for gateway(KGW)
// authentication. It's the parameters that will be used to compose the
// message(SIWE like) to sign.
type gatewayAuthParameter struct {
	Nonce          string `json:"nonce"`
	Statement      string `json:"statement"` // optional
	IssueAt        string `json:"issue_at"`
	ExpirationTime string `json:"expiration_time"`
}

// gatewayAuthGetResponse defines the response of GET request for gateway(KGW) authentication.
type gatewayAuthGetResponse struct {
	Result *gatewayAuthParameter `json:"result"`
	Error  string                `json:"error"`
}

// requestGatewayAuthParameter requests authentication parameters from the gateway(KGW) provider.
// This will send a GET request to KGW provider, and the provider will return
// parameters that the user could use to compose a message to sign.
func requestGatewayAuthParameter(hc *http.Client, target string) (*gatewayAuthParameter, error) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request authn: %w", err)
	}
	defer resp.Body.Close()

	var r gatewayAuthGetResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if r.Error != "" {
		return nil, fmt.Errorf("got error response: %s", r.Error)
	}

	return r.Result, nil
}

// composeGatewayAuthMessage composes the SIWE-like message to sign.
// param is the result of GET request for authentication.
// ALl the other parameters are expected from config.
func composeGatewayAuthMessage(param *gatewayAuthParameter, domain string, uri string,
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

type gatewayAuthPostResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// gatewayAuthPostPayload defines the payload of POST request for authentication
type gatewayAuthPostPayload struct {
	Nonce     string          `json:"nonce"`  // identifier for authn session
	Sender    []byte          `json:"sender"` // sender public key
	Signature *auth.Signature `json:"signature"`
}

// requestGatewayAuthCookie requests an authenticated cookie from the gateway(KGW) provider.
// This will send a POST request to the KGW provider, and the provider will
// return a cookie.
func requestGatewayAuthCookie(hc *http.Client, target string, nonce string,
	sender []byte, sig *auth.Signature) (*http.Cookie, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	authURL, err := targetURL.Parse(kgwAuthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse authURL: %w", err)
	}

	payload := gatewayAuthPostPayload{
		Nonce:     nonce,
		Signature: sig,
		Sender:    sender,
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(&payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, authURL.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request authn: %w", err)
	}
	defer resp.Body.Close()

	var r gatewayAuthPostResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if r.Error != "" {
		return nil, fmt.Errorf("error response: %s", r.Error)
	}

	// NOTE: one cookie is enough for now
	for _, c := range resp.Cookies() {
		if c.Name == kgwAuthCookieName {
			return c, nil
		}
	}

	return nil, fmt.Errorf("no cookie returned")
}
