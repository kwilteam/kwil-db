package client

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The following errors may be detected by consumers using errors.Is.
var (
	// ErrUnauthorized is returned when the client is not authenticated
	// It is the equivalent of http status code 401
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
)

// ConvertGRPCErr will convert the error to a known type, if possible.
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

// RPCError is a common error type used by any RPC client implementation to
// provide a detectable error to consumers using errors.As. We may define our
// own codes. Instances of RPCError may also be combined with other error types
// defined above using errors.Join.
type RPCError struct {
	Msg  string
	Code int32
}

func (err RPCError) Error() string {
	return fmt.Sprintf("err code = %d, msg = %v", err.Code, err.Msg)
}
