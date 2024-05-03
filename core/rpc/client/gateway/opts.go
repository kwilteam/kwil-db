package gateway

import (
	"net/http"
	"net/http/cookiejar"
)

type ClientOption func(*clientOptions)

// WithHTTPClient sets the http client for the client.
// This allows custom http clients to be used.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *clientOptions) {
		c.Client = client
	}
}

type clientOptions struct {
	Client *http.Client
}

// DefaultClientOptions returns the default client options, which ensure the
// connection has a cookie jar.
func DefaultClientOptions() *clientOptions {
	jar, _ := cookiejar.New(nil)
	return &clientOptions{
		Client: &http.Client{
			Jar: jar,
		},
	}
}
