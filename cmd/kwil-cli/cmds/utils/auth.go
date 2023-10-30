package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	httpRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
)

const (
	kgwAuthEndpoint   = "/auth"
	kgwAuthCookieName = "kgw_session"
)

var authCmdDesc = `Authentication is required to use the KGW provider. This command will
prompt for a signature and return a token for later.
KGW authentication is not part of Kwild API.
`

// authCmd is the command to authenticate to a KGW provider.
func authCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "auth",
		Short: "Auth is used to authenticate to a KGW provider", // or sass provider?
		Long:  authCmdDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := config.LoadCliConfig()
			if err != nil {
				return err
			}

			if conf.PrivateKey == nil {
				return fmt.Errorf("private key not provided")
			}

			// TODO: verify chainID and other info
			signer := auth.EthPersonalSigner{Key: *conf.PrivateKey}
			if conf.GrpcURL == "" {
				// this is somewhat redundant since the config marks it as required, but in case the config is changed
				return fmt.Errorf("provider url not provided")
			}

			userAddress, err := signer.Address()
			if err != nil {
				return fmt.Errorf("get address: %w", err)
			}

			userPubkey := signer.PublicKey()

			// TODO: if already exists, prompt to overwrite ?????

			// KGW auth is not part of Kwil API, we use a standard http client
			// this client is to reuse connection
			hc := httpRPC.DefaultJarHTTPClient()
			msg, err := requestForAuthentication(hc, conf.GrpcURL, userAddress)
			if err != nil {
				return fmt.Errorf("request for authentication: %w", err)
			}

			sig, err := promptSigning(&signer, msg)
			if err != nil {
				return fmt.Errorf("prompt signing: %w", err)
			}

			token, err := requestForToken(hc, conf.GrpcURL, userPubkey, sig)
			if err != nil {
				return fmt.Errorf("request for token: %w", err)
			}

			fmt.Println("cookie:", token.String())

			err = common.SaveAuthInfo(common.KGWAuthTokenFilePath(), userAddress, token)
			if err != nil {
				return fmt.Errorf("save auth token: %w", err)
			}

			// NOTE: seems no point support JSON output format?
			fmt.Println("Authentication successful")
			return nil
		},
	}

	return cmd
}

// requestForAuthentication requests authentication message from the KGW provider.
// This will send a GET request to KGW provider, and the provider will return
// a message that the user needs to check and sign.
// NOTE: this workflow will change to the same workflow in web browser, for now
// we just sign whatever the provider returns for quick prototyping.
// The ideal workflow will be:
// 1. cli requests authentication
// 2. KGW provider returns a nonce to cli
// 3. cli composes a message with the nonce and sign it, send to KGW provider
// 4. KGW provider verifies the signature and returns a token
func requestForAuthentication(hc *http.Client, target string, address string) (string, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return "", fmt.Errorf("parse target: %w", err)
	}

	params := url.Values{
		"from": []string{address},
	}

	authURL, err := targetURL.Parse(kgwAuthEndpoint + "?" + params.Encode())
	if err != nil {
		return "", fmt.Errorf("parse authURL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, authURL.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var r authResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if r.Error != "" {
		return "", fmt.Errorf("got error response: %s", r.Error)
	}

	return r.Result, nil
}

// promptSigning prompts the user to sign a message. User should be aware of
// the message content, since in this workflow user will only see the message once.
func promptSigning(signer auth.Signer, msg string) (*auth.Signature, error) {
	// display the message to user
	fmt.Println(msg)

	prompt := promptui.Prompt{
		Label:     "Do you want to sign this message?",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("you declined to sign")
	}

	return signer.Sign([]byte(msg))
}

type authResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

type authPostPayload struct {
	Sender    []byte          `json:"sender"` // sender public key
	Signature *auth.Signature `json:"signature"`
}

// requestForToken requests a token from the KGW provider
// This will send a POST request to the KGW provider, and the provider will return a token in cookie header
func requestForToken(hc *http.Client, target string, sender []byte,
	sig *auth.Signature) (*http.Cookie, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	authURL, err := targetURL.Parse(kgwAuthEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse authURL: %w", err)
	}

	payload := authPostPayload{
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
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var r authResponse
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
