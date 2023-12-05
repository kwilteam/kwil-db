// package gatewayclient implements a client for kwild that can also authenticate
// with a kwil gateway.
package gatewayclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/kwilteam/kwil-db/core/client"
	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/gateway"
	httpGateway "github.com/kwilteam/kwil-db/core/rpc/client/gateway/http"
	httpTx "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	gatewayTypes "github.com/kwilteam/kwil-db/core/types/gateway"
)

// GatewayClient is a client that is made to interact with a kwil gateway.
// It inherits the functionality of the main Kwil client, but also provides
// authentication cookies to the gateway.
// It automatically handles the authentication process with the gateway.
type GatewayClient struct {
	client.Client

	target *url.URL

	httpClient    *http.Client
	gatewayClient gateway.GatewayClient

	gatewaySigner GatewayAuthSignFunc // a hook for when the gateway authentication is needed
}

func NewClient(ctx context.Context, target string, opts *GatewayOptions) (*GatewayClient, error) {
	options := DefaultOptions()
	options.Apply(opts)

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	httpClient := &http.Client{
		Jar: cookieJar,
	}

	parsedTarget, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}

	txClient := httpTx.NewClient(parsedTarget, httpTx.WithHTTPClient(httpClient))

	coreClient, err := client.WrapClient(ctx, txClient, &options.ClientOptions)
	if err != nil {
		return nil, fmt.Errorf("wrap client: %w", err)
	}

	gatewayRPC, err := httpGateway.NewGatewayHttpClient(parsedTarget, httpGateway.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create gateway rpc client: %w", err)
	}

	g := &GatewayClient{
		Client:        *coreClient,
		httpClient:    httpClient,
		gatewaySigner: options.AuthSignFunc,
		gatewayClient: gatewayRPC,
		target:        parsedTarget,
	}

	return g, nil
}

// CallAction call an action. It returns the result records.  If authentication is needed,
// it will call the gatewaySigner to sign the authentication message.
func (c *GatewayClient) CallAction(ctx context.Context, dbid string, action string, inputs []any) (*client.Records, error) {
	// we will try to call with the current cookies set.  If we receive an error and it is an auth error,
	// we will re-auth and retry.  We will only retry once.
	res, err := c.Client.CallAction(ctx, dbid, action, inputs)
	if err == nil {
		return res, nil
	}
	if !errors.Is(err, rpcClient.ErrUnauthorized) {
		return nil, err
	}

	// we need to authenticate
	err = c.authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("authenticate error: %w", err)
	}

	// retry the call
	return c.Client.CallAction(ctx, dbid, action, inputs)
}

// authenticate authenticates the client with the gateway.
func (c *GatewayClient) authenticate(ctx context.Context) error {
	authParam, err := c.gatewayClient.GetAuthParameter(ctx)
	if err != nil {
		return fmt.Errorf("get auth parameter: %w", err)
	}

	authURI, err := url.JoinPath(c.target.String(), gateway.AuthEndpoint)
	if err != nil {
		return fmt.Errorf("join path: %w", err)
	}

	msg := composeGatewayAuthMessage(authParam, c.target.String(), authURI, kgwAuthVersion, c.Client.ChainID())

	if c.Signer == nil {
		return fmt.Errorf("cannot authenticate to gateway without a signer")
	}
	sig, err := c.gatewaySigner(msg, c.Signer)
	if err != nil {
		return fmt.Errorf("sign message: %w", err)
	}

	// send the auth request
	return c.gatewayClient.Auth(ctx, &gatewayTypes.GatewayAuth{
		Nonce:     authParam.Nonce,
		Sender:    c.Signer.Identity(),
		Signature: sig,
	})
}

// GetAuthCookie returns the authentication cookie currently being used.
// If no authentication cookie is being used, it returns nil, false.
func (c *GatewayClient) GetAuthCookie() (cookie *http.Cookie, found bool) {
	cookies := c.httpClient.Jar.Cookies(c.target)
	for _, cookie := range cookies {
		if cookie.Name == kgwAuthCookieName {
			return cookie, true
		}
	}
	return nil, false
}

// SetAuthCookie sets the authentication cookie to be used.
// If the cookie is not valid for the client target, it returns an error.
// It will overwrite any existing authentication cookie.
func (c *GatewayClient) SetAuthCookie(cookie *http.Cookie) error {
	if cookie.Domain != "" && cookie.Domain != c.target.Host {
		return fmt.Errorf("cookie domain %s not valid for host %s", cookie.Domain, c.target.Host)
	}
	if cookie.Name != kgwAuthCookieName {
		return fmt.Errorf("cookie name %s not valid", cookie.Name)
	}

	c.httpClient.Jar.SetCookies(c.target, []*http.Cookie{cookie})

	return nil
}

// Authenticate will authenticate the client with the gateway.
// This is not necessary, as the client will automatically authenticate when needed,
// however it can be useful if the client desires to control when the authentication
// occurs / wants to manually force re-authentication.
func (c *GatewayClient) Authenticate(ctx context.Context) error {
	return c.authenticate(ctx)
}
