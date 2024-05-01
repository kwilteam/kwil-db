package rpcserver

import (
	"context"
	"encoding/json"

	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
)

// MethodHandler is a type of function that returns an interface containing a
// pointer to a handler's input arguments, and a handler function that captures
// the arguments pointer. The handler function returns its result type in an
// interface, and a *jsonrpc.Error. A simple MethodHandler would instantiate a
// new concrete instance of the parameters type and define a function that uses
// that instance to perform some operations.
type MethodHandler func(ctx context.Context, s *Server) (argsPtr any, handler func() (any, *jsonrpc.Error))

// Svc is a type that enumerates its handler functions by method name. To
// handle a method, the Server:
//  1. retrieves the MethodHandler associated with the method
//  2. calls the MethodHandler to get the args interface and handler function
//  3. unmarshals the inputs from a json.RawMessage into the args interface
//  4. calls the handler function, returning the result and error
//  5. marshal either the result or the Error into a Response
type Svc interface {
	Handlers() map[jsonrpc.Method]MethodHandler
}

// RegisterSvc registers every MethodHandler for a service.
//
// The Server's fixed endpoint is used.
func (s *Server) RegisterSvc(svc Svc) {
	for method, handler := range svc.Handlers() {
		s.log.Debugf("Registering method %q", method)
		s.RegisterMethodHandler(method, handler)
	}
}

// RegisterMethodHandler registers a single MethodHandler.
// See also RegisterSvc.
func (s *Server) RegisterMethodHandler(method jsonrpc.Method, h MethodHandler) {
	s.methodHandlers[method] = h
}

// handleMethod unmarshals into the appropriate params struct, and dispatches to
// the TxSvc method handler.
func (s *Server) handleMethod(ctx context.Context, method jsonrpc.Method, params json.RawMessage) (any, *jsonrpc.Error) {
	maker, have := s.methodHandlers[method]
	if !have {
		return nil, jsonrpc.NewError(jsonrpc.ErrorUnknownMethod, "unknown method", nil)
	}

	argsPtr, handler := maker(ctx, s)

	err := json.Unmarshal(params, argsPtr)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, err.Error(), nil)
	}

	return handler()
}
