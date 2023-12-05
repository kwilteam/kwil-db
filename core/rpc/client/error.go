package client

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrUnauthorized is returned when the client is not authenticated
	// It is the equivalent of http status code 401
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// convertErr will convert the error to a known type, if possible.
// It is expected that the error is from a gRPC call.
func ConvertGRPCErr(err error) error {
	statusError, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("unrecognized error: %w", err)
	}

	switch statusError.Code() {
	case codes.OK:
		// this should never happen?
		return fmt.Errorf("unexpected OK status code returned error")
	case codes.NotFound:
		return ErrNotFound
	}

	return fmt.Errorf("%v (%d)", statusError.Message(), statusError.Code())
}
