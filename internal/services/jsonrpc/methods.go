package rpcserver

import (
	"context"
	"encoding/json"
	"reflect"

	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/openrpc"
)

// MethodHandler is a type of function that returns an interface containing a
// pointer to a handler's input arguments, and a handler function that captures
// the arguments pointer. The handler function returns its result type in an
// interface, and a *jsonrpc.Error. A simple MethodHandler would instantiate a
// new concrete instance of the parameters type and define a function that uses
// that instance to perform some operations.
type MethodHandler func(ctx context.Context, s *Server) (argsPtr any, handler func() (any, *jsonrpc.Error))

type Handler[I, O any] func(context.Context, *I) (*O, *jsonrpc.Error)

func ioTypes[I, O any](Handler[I, O]) (reflect.Type, reflect.Type) {
	return reflect.TypeOf((*I)(nil)).Elem(), reflect.TypeOf((*O)(nil)).Elem() // reflect.TypeFor[I](), reflect.TypeFor[O]()
}

func MakeMethodHandler[I, O any](fn Handler[I, O]) MethodHandler {
	return func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := new(I)
		return req, func() (any, *jsonrpc.Error) { return fn(ctx, req) }
	}
}

func InspectHandler[I, O any](fn Handler[I, O]) (reflect.Type, reflect.Type, MethodHandler) {
	reqType, respType := ioTypes(fn)
	return reqType, respType, MakeMethodHandler(fn)
}

type MethodDef struct {
	Desc       string
	ParamDescs []string
	RespDesc   string
	Handler    MethodHandler
	ReqType    reflect.Type
	RespType   reflect.Type
}

func MakeMethodDef[I, O any](handler Handler[I, O], desc, respDesc string) MethodDef {
	iT, oT, rpcHandler := InspectHandler(handler)
	return MethodDef{
		Desc:     desc,
		RespDesc: respDesc,
		Handler:  rpcHandler,
		ReqType:  iT,
		RespType: oT,
	}
}

// Svc is a type that enumerates its handler functions by method name. To
// handle a method, the Server:
//  1. retrieves the MethodHandler associated with the method
//  2. calls the MethodHandler to get the args interface and handler function
//  3. unmarshals the inputs from a json.RawMessage into the args interface
//  4. calls the handler function, returning the result and error
//  5. marshal either the result or the Error into a Response
type Svc interface {
	Methods() map[jsonrpc.Method]MethodDef
}

// RegisterSvc registers every MethodHandler for a service.
//
// The Server's fixed endpoint is used.
func (s *Server) RegisterSvc(svc Svc) {
	for method, def := range svc.Methods() {
		s.log.Debugf("Registering method %q", method)
		s.RegisterMethodHandler(method, def.Handler)
		s.methodDefs[string(method)] = &openrpc.MethodDefinition{
			Description:  def.Desc,
			RequestType:  def.ReqType,
			ResponseType: def.RespType,
			RespTypeDesc: def.RespDesc,
		}
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
