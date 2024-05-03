package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync/atomic"

	rpcClient "github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/client/gateway"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
)

type Client struct {
	conn     *http.Client
	endpoint string

	reqID atomic.Int64
}

// NewClient creates a new gateway json rpc client, target should be the base URL
// of the gateway server, and should not include "/rpc/v1" as that is appended
// automatically. If the client does not have a cookie jar, an error is returned.
func NewClient(target *url.URL, opts ...gateway.ClientOption) (*Client, error) {
	// This client uses API v1 methods and request/response types.
	target = target.JoinPath("/rpc/v1")

	c := gateway.DefaultClientOptions()
	for _, o := range opts {
		o(c)
	}

	g := &Client{
		conn:     c.Client,
		endpoint: target.String(),
	}

	// if the caller passed a custom http client without a cookie jar, return an error
	if g.conn.Jar == nil {
		return nil, errors.New("gateway http client must have a cookie jar")
	}

	return g, nil
}

var _ gateway.Client = (*Client)(nil)

func (g *Client) nextReqID() string {
	id := g.reqID.Add(-1) // Decrement by 1, so it won't be duplicated with txClient
	return strconv.FormatInt(id, 10)
}

// Authn authenticates the client with the gateway.
// It sets the returned cookie in the client's cookie jar.
func (g *Client) Authn(ctx context.Context, auth *gateway.AuthnRequest) error {
	res := &gateway.AuthnResponse{}
	err := g.call(ctx, gateway.MethodAuthn, auth, res)
	return err
}

// GetAuthnParameter returns the auth parameter for the client.
func (g *Client) GetAuthnParameter(ctx context.Context) (*gateway.AuthnParameterResponse, error) {
	res := &gateway.AuthnParameterResponse{}
	err := g.call(ctx, gateway.MethodAuthnParam, &gateway.AuthnParameterRequest{}, res)
	return res, err
}

// call sends a JSON-RPC request to the gateway.
// NOTE: this is basically a copy of core/rpc/client/user/jsonrpc/client.call
func (cl *Client) call(ctx context.Context, method string, cmd, res any) error {
	// res needs to be a pointer otherwise we can't unmarshal into it.
	if rtp := reflect.TypeOf(res); rtp.Kind() != reflect.Ptr {
		return errors.New("result must be a pointer")
	}

	// Marshal the params.
	params, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	// Marshal the request.
	id := cl.nextReqID()
	req := jsonrpc.NewRequest(id, method, params)

	request, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Build and perform the http request.
	requestReader := bytes.NewReader(request)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		cl.endpoint, requestReader)
	if err != nil {
		return fmt.Errorf("failed to construct new http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// httpReq.SetBasicAuth(c.User, c.Pass)

	httpResponse, err := cl.conn.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}
	defer httpResponse.Body.Close()

	// For the most part we ignore the http status code in favor of structured
	// errors in the response, but in case we cannot decode any response body,
	// get an error based on the http status code.
	var httpErr error
	switch status := httpResponse.StatusCode; status {
	case http.StatusOK: // expected with nil resp.Error
	case http.StatusUnauthorized:
		httpErr = rpcClient.ErrUnauthorized
	case http.StatusNotFound:
		httpErr = rpcClient.ErrNotFound
	case http.StatusInternalServerError:
		httpErr = errors.New("server error")
	default:
		if status >= 400 {
			httpErr = errors.New(http.StatusText(status))
		}
	}

	resp := &jsonrpc.Response{}
	err = json.NewDecoder(httpResponse.Body).Decode(resp)
	if err != nil {
		if httpErr != nil {
			return httpErr
		}
		return fmt.Errorf("failed to decode response: %w", errors.Join(err, httpErr))
	}

	if resp.Error != nil {
		return clientError(resp.Error)
	} // any not OK http status code should have

	if resp.JSONRPC != "2.0" { // indicates response body was not a jsonrpc.Response but didn't fail Decode
		if httpErr != nil {
			return httpErr
		}
		return fmt.Errorf("invalid JSON-RPC response")
	}

	if err = json.Unmarshal(resp.Result, res); err != nil {
		return fmt.Errorf("failed to decode result as response: %w", errors.Join(err, httpErr))
	}

	return nil
}

// clientError joins a jsonrpc.Error with a client.RPCError and any appropriate
// named error kind like ErrNotFound, ErrUnauthorized, etc. based on the code.
func clientError(jsonRPCErr *jsonrpc.Error) error {
	// TODO: make clientError from core/rpc/client/user/jsonrpc/client.clientError reusable
	// Then we will first check if we recognize the error(KGW) code, if not,
	// we use client.clientError

	rpcErr := &rpcClient.RPCError{
		Msg:  jsonRPCErr.Message,
		Code: int32(jsonRPCErr.Code),
	}
	err := errors.Join(jsonRPCErr, rpcErr)

	switch jsonRPCErr.Code {
	case jsonrpc.ErrorKGWNotAuthorized:
		return errors.Join(rpcClient.ErrUnauthorized, err)
	case jsonrpc.ErrorKGWInvalidPayload, jsonrpc.ErrorInvalidParams,
		jsonrpc.ErrorParse, jsonrpc.ErrorInvalidRequest:
		return errors.Join(rpcClient.ErrInvalidRequest, err)
	case jsonrpc.ErrorKGWNotFound, jsonrpc.ErrorUnknownMethod:
		return errors.Join(rpcClient.ErrNotFound, err)
	case jsonrpc.ErrorKGWNotAllowed:
		return errors.Join(rpcClient.ErrNotAllowed, err)
	default:
	}

	return err
}
