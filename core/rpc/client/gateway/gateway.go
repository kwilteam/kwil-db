package gateway

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types/gateway"
)

// GatewayClient is the interface for the Kwil gateway client.
type GatewayClient interface {
	// Auth sends an authentication request to the gateway.
	// It will set a cookie in the client if successful.
	Auth(ctx context.Context, req *gateway.GatewayAuth) error

	// GetAuthParameter returns the parameters that will be used to compose the
	// authentication message that will be signed by the client.
	GetAuthParameter(ctx context.Context) (*gateway.GatewayAuthParameter, error)

	// TODO: how do we check for cookie expiration?
}

var (
	// ErrAuthenticationFailed is returned when the gateway authentication fails.
	// This is usually due to an incorrect signature.
	ErrAuthenticationFailed = fmt.Errorf("gateway authentication failed")

	// ErrGatewayNotAuthenticated is returned when the client is not authenticated
	// with the gateway.
	ErrGatewayNotAuthenticated = fmt.Errorf("gateway not authenticated")
)

const (
	// AuthEndpoint is the endpoint for the gateway authentication.
	// This is defined here since the endpoint is part of the signature
	// protocol, which both the gateway and client are expected to follow.
	AuthEndpoint = "/auth"
)
