package http

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	httpFunction "github.com/kwilteam/kwil-db/core/rpc/http/function"
)

var ErrInvalidSignature = errors.New("invalid signature")

type Client struct {
	conn *httpFunction.APIClient
	url  *url.URL
}

// NewClient creates a new http client for the Kwil user service
func NewClient(target *url.URL, opts ...ClientOption) *Client {
	c := &Client{
		url: target,
	}

	clientOpts := &clientOptions{
		client: &http.Client{},
	}

	for _, o := range opts {
		o(clientOpts)
	}

	cfg := httpFunction.NewConfiguration()
	cfg.HTTPClient = clientOpts.client
	cfg.BasePath = strings.TrimRight(target.String(), "/")
	cfg.Host = target.Host
	cfg.Scheme = target.Scheme

	c.conn = httpFunction.NewAPIClient(cfg)

	return c
}

func (c *Client) VerifySignature(ctx context.Context, sender []byte, signature *auth.Signature, message []byte) error {
	result, res, err := c.conn.FunctionServiceApi.FunctionServiceVerifySignature(ctx, httpFunction.FunctionVerifySignatureRequest{
		Sender: base64.StdEncoding.EncodeToString(sender),
		Msg:    base64.StdEncoding.EncodeToString(message),
		Signature: &httpFunction.TxSignature{
			SignatureBytes: base64.StdEncoding.EncodeToString(signature.Signature),
			SignatureType:  signature.Type,
		},
	})
	if err != nil { // communication error
		return err
	}
	defer res.Body.Close()

	// server logic error
	if result.Error_ != "" {
		return fmt.Errorf("%w: %s", ErrInvalidSignature, result.Error_)
	}

	// NOTE: Forget why I put both `valid` and `error` in the response.
	// if `valid` is false, `error` should not be empty.
	// This might be not needed, but I just keep it here.
	if !result.Valid {
		return ErrInvalidSignature
	}

	return nil
}
