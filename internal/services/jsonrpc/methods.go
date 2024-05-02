package rpcserver

import (
	"context"
	"encoding/json"

	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
)

// This file defines the method dispatcher to the TxSvc methods.

// handleMethod0 unmarshals into the appropriate params struct, and dispatches to
// the TxSvc method handler.
func handleMethod0(ctx context.Context, s *Server, method jsonrpc.Method, params json.RawMessage) (any, *jsonrpc.Error) { //nolint:unused
	var args any // assign a pointer to a concrete type for unmarshal
	var handler func() (any, *jsonrpc.Error)

	switch method { // see handlerMethod below for a hashmap solution
	case jsonrpc.MethodAccount:
		req := &jsonrpc.AccountRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Account(ctx, req) }
	case jsonrpc.MethodBroadcast:
		req := &jsonrpc.BroadcastRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Broadcast(ctx, req) }
	case jsonrpc.MethodCall:
		req := &jsonrpc.CallRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Call(ctx, req) }
	case jsonrpc.MethodChainInfo:
		req := &jsonrpc.ChainInfoRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.ChainInfo(ctx, req) }
	case jsonrpc.MethodDatabases:
		req := &jsonrpc.ListDatabasesRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.ListDatabases(ctx, req) }
	case jsonrpc.MethodPing:
		req := &jsonrpc.PingRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Ping(ctx, req) }
	case jsonrpc.MethodPrice:
		req := &jsonrpc.EstimatePriceRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.EstimatePrice(ctx, req) }
	case jsonrpc.MethodQuery:
		req := &jsonrpc.QueryRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Query(ctx, req) }
	case jsonrpc.MethodSchema:
		req := &jsonrpc.SchemaRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.Schema(ctx, req) }
	case jsonrpc.MethodTxQuery:
		req := &jsonrpc.TxQueryRequest{}
		args, handler = req, func() (any, *jsonrpc.Error) { return s.svc.TxQuery(ctx, req) }
	default:
		return nil, jsonrpc.NewError(jsonrpc.ErrorUnknownMethod, "unknown method", nil)
	}

	err := json.Unmarshal(params, args)
	if err != nil {
		return nil, jsonrpc.NewError(jsonrpc.ErrorInvalidParams, err.Error(), nil)
	}

	return handler()
}

// hashmap solution below.

// handleMethod unmarshals into the appropriate params struct, and dispatches to
// the TxSvc method handler.
func handleMethod(ctx context.Context, s *Server, method jsonrpc.Method, params json.RawMessage) (any, *jsonrpc.Error) {
	maker, have := methodHandlers[method]
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

type handlerMaker func(ctx context.Context, s *Server) (argsPtr any, handler func() (any, *jsonrpc.Error))

var methodHandlers = map[jsonrpc.Method]handlerMaker{
	jsonrpc.MethodAccount: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.AccountRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Account(ctx, req) }
	},
	jsonrpc.MethodBroadcast: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.BroadcastRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Broadcast(ctx, req) }
	},
	jsonrpc.MethodCall: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.CallRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Call(ctx, req) }
	},
	jsonrpc.MethodChainInfo: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.ChainInfoRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.ChainInfo(ctx, req) }
	},
	jsonrpc.MethodDatabases: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.ListDatabasesRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.ListDatabases(ctx, req) }
	},
	jsonrpc.MethodPing: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.PingRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Ping(ctx, req) }
	},
	jsonrpc.MethodPrice: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.EstimatePriceRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.EstimatePrice(ctx, req) }
	},
	jsonrpc.MethodQuery: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.QueryRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Query(ctx, req) }
	},
	jsonrpc.MethodSchema: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.SchemaRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.Schema(ctx, req) }
	},
	jsonrpc.MethodTxQuery: func(ctx context.Context, s *Server) (any, func() (any, *jsonrpc.Error)) {
		req := &jsonrpc.TxQueryRequest{}
		return req, func() (any, *jsonrpc.Error) { return s.svc.TxQuery(ctx, req) }
	},
}
