package client

import "errors"

var (
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrUnauthorized is returned when the client is not authenticated
	// It is the equivalent of http status code 401
	ErrUnauthorized = errors.New("unauthorized")
)
