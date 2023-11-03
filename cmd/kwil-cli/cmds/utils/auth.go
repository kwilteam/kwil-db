package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	httpRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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
			hc := httpRPC.DefaultHTTPClient()

			authURI, err := url.JoinPath(conf.GrpcURL, kgwAuthEndpoint)
			if err != nil {
				panic(err)
			}

			authParam, err := requestAuthParameter(hc, authURI, userAddress)
			if err != nil {
				return fmt.Errorf("request for authentication: %w", err)
			}

			// TODO: need a authVersion configuration, or should this be part of the authParam?
			// It seems reasonable to have this as part of the authParam, and SDK
			// will use the version to compose the message using different template.
			// According to SIWE, the version is fixed to 1 right now.
			msg := composeAuthMessage(authParam, conf.GrpcURL, userAddress,
				authURI, "1", conf.ChainID)

			sig, err := promptSigning(&signer, msg)
			if err != nil {
				return fmt.Errorf("prompt signing: %w", err)
			}

			token, err := requestAuthToken(hc, conf.GrpcURL, userPubkey, sig)
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

// requestAuthParameter requests authentication parameters from the KGW provider.
// This will send a GET request to KGW provider, and the provider will return
// parameters that the user could use to compose a message to sign.
// NOTE: this workflow will change to the same workflow in web browser, for now
// we just sign whatever the provider returns for quick prototyping.
// The ideal workflow will be:
// 1. cli requests authentication
// 2. KGW provider returns a nonce to cli
// 3. cli composes a message with the nonce and sign it, send to KGW provider
// 4. KGW provider verifies the signature and returns a token
func requestAuthParameter(hc *http.Client, target string, address string) (*authParam, error) {
	params := url.Values{
		"from": []string{address},
	}

	req, err := http.NewRequest(http.MethodGet, target+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var r authGetResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if r.Error != "" {
		return nil, fmt.Errorf("got error response: %s", r.Error)
	}

	return r.Result, nil
}

// composeAuthMessage composes the SIWE-like message to sign.
// param is the result of GET request for authentication.
// ALl the other parameters are expected from config.
func composeAuthMessage(param *authParam, domain string, address string, uri string, version string, chainID string) string {
	var msg bytes.Buffer
	msg.WriteString(
		fmt.Sprintf("%s wants you to sign in with your account:\n", domain))
	msg.WriteString(fmt.Sprintf("%s\n", address))
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

// authParam defines the result of GET request for authentication.
// It's the parameters that will be used to compose the message(SIWE like)
// to sign.
type authParam struct {
	Nonce          string `json:"nonce"`
	Statement      string `json:"statement"` // optional
	IssueAt        string `json:"issue_at"`
	ExpirationTime string `json:"expiration_time"`
}

type authGetResponse struct {
	Result *authParam `json:"result"`
	Error  string     `json:"error"`
}

type authPostResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

// authPostPayload defines the payload of POST request for authentication
type authPostPayload struct {
	Sender    []byte          `json:"sender"` // sender public key
	Signature *auth.Signature `json:"signature"`
}

// requestAuthToken requests a authenticated token from the KGW provider.
// This will send a POST request to the KGW provider, and the provider will
// return a cookie.
func requestAuthToken(hc *http.Client, target string, sender []byte,
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

	var r authPostResponse
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
