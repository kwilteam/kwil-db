package http2

import (
	"net/http"
)

type ClientOption func(*clientOptions)

// WithHTTPClient sets the http client for the client.
// This allows custom http clients to be used.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientOptions) {
		c.client = client
	}
}

type clientOptions struct {
	client *http.Client
}
