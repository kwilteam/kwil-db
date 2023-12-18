package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/gateway"
	types "github.com/kwilteam/kwil-db/core/types/gateway"
)

type GatewayHttpClient struct {
	client *http.Client
	target *url.URL
}

// NewGatewayHttpClient creates a new gateway http client.
// If the client does not have a cookie jar, an error is returned.
func NewGatewayHttpClient(target *url.URL, opts ...ClientOption) (*GatewayHttpClient, error) {
	// gateway needs access to the cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &clientOptions{
		client: &http.Client{
			Jar: jar,
		},
	}

	for _, o := range opts {
		o(c)
	}

	g := &GatewayHttpClient{
		client: c.client,
		target: target,
	}

	// if the caller passed a custom http client without a cookie jar, return an error
	// otherwise, it will not error in the future- it simply will not work as expected
	if g.client.Jar == nil {
		return nil, errors.New("gateway http client must have a cookie jar")
	}

	return g, nil
}

var _ gateway.GatewayClient = (*GatewayHttpClient)(nil)

// fullUrl returns the full url for the given endpoint
func (g *GatewayHttpClient) fullUrl(endpoint string) string {
	return g.target.JoinPath(endpoint).String()
}

// Auth authenticates the client with the gateway.
// It sets the returned cookie in the client's cookie jar.
func (g *GatewayHttpClient) Auth(ctx context.Context, auth *types.GatewayAuth) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(auth)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.fullUrl(gateway.AuthEndpoint), buf)
	if err != nil {
		return err
	}

	res, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Join(rpcClient.ErrNotFound, errors.New(res.Status))
	}

	var r gatewayAuthPostResponse
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return err
	}

	if r.Error != "" {
		return errors.New(r.Error)
	}

	return nil
}

// GetAuthParameter returns the auth parameter for the client.
func (g *GatewayHttpClient) GetAuthParameter(ctx context.Context) (*types.GatewayAuthParameter, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.fullUrl(gateway.AuthEndpoint), nil)
	if err != nil {
		return nil, err
	}

	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.Join(rpcClient.ErrNotFound, errors.New(res.Status))
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	var r gatewayAuthGetResponse
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.Error != "" {
		return nil, errors.New(r.Error)
	}

	return r.Result, nil
}

// gatewayAuthPostResponse defines the response of POST request for /auth
type gatewayAuthPostResponse struct {
	Error string `json:"error"`
}

// gatewayAuthGetResponse defines the response of GET request for /auth
type gatewayAuthGetResponse struct {
	Result *types.GatewayAuthParameter `json:"result"`
	Error  string                      `json:"error"`
}
