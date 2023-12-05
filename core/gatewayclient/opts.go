package gatewayclient

import (
	"github.com/kwilteam/kwil-db/core/client"
)

// GatewayOptions are options that can be set for the gateway client
type GatewayOptions struct {
	client.ClientOptions

	// AuthSignFunc is a function that will be used to sign gateway authentication messages.
	AuthSignFunc GatewayAuthSignFunc
}

// DefaultOptions returns the default options for the gateway client.
func DefaultOptions() *GatewayOptions {
	return &GatewayOptions{
		ClientOptions: *client.DefaultOptions(),

		AuthSignFunc: defaultGatewayAuthSignFunc,
	}
}

// Apply applies the passed options to the receiver.
func (c *GatewayOptions) Apply(opt *GatewayOptions) {
	if opt == nil {
		return
	}

	c.ClientOptions.Apply(&opt.ClientOptions)

	if opt.AuthSignFunc != nil {
		c.AuthSignFunc = opt.AuthSignFunc
	}
}
