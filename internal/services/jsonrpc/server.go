package rpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
)

var logger log.Logger

// Server is a JSON-RPC server for the Kwil "tx" service.
type Server struct {
	srv *http.Server
	log log.Logger
	svc TxSvc
	// TODO: maybe separate the HTTP serving pieces from the dispatching to
	// service methods. Instead of depending on a TxSvc, have externally
	// registered method handlers that use the service.

	// methodHandlers map[jsonrpc.Method]MethodHandler
}

type TxSvc interface {
	ChainInfo(context.Context, *jsonrpc.ChainInfoRequest) (*jsonrpc.ChainInfoResponse, *jsonrpc.Error)
	Broadcast(context.Context, *jsonrpc.BroadcastRequest) (*jsonrpc.BroadcastResponse, *jsonrpc.Error)
	EstimatePrice(context.Context, *jsonrpc.EstimatePriceRequest) (*jsonrpc.EstimatePriceResponse, *jsonrpc.Error)
	Query(context.Context, *jsonrpc.QueryRequest) (*jsonrpc.QueryResponse, *jsonrpc.Error)
	Account(context.Context, *jsonrpc.AccountRequest) (*jsonrpc.AccountResponse, *jsonrpc.Error)
	Ping(context.Context, *jsonrpc.PingRequest) (*jsonrpc.PingResponse, *jsonrpc.Error)
	ListDatabases(context.Context, *jsonrpc.ListDatabasesRequest) (*jsonrpc.ListDatabasesResponse, *jsonrpc.Error)
	Schema(context.Context, *jsonrpc.SchemaRequest) (*jsonrpc.SchemaResponse, *jsonrpc.Error)
	Call(context.Context, *jsonrpc.CallRequest) (*jsonrpc.CallResponse, *jsonrpc.Error)
	TxQuery(context.Context, *jsonrpc.TxQueryRequest) (*jsonrpc.TxQueryResponse, *jsonrpc.Error)
}

// NewServer creates a new JSON-RPC server. Presently this requires a TxSvc, but
// it should switch to externally registered routes.
func NewServer(addr string, log log.Logger, svc TxSvc) (*Server, error) {
	mux := http.NewServeMux() // http.DefaultServeMux has the pprof endpoints mounted
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}

	s := &Server{
		srv: srv,
		log: log,
		svc: svc,
	}

	mux.Handle("/rpc/v1", http.HandlerFunc(s.handlerV1))

	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	s.log.Info("JSON-RPC server listening", log.String("address", ln.Addr().String()))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.srv.Serve(ln)
		if !errors.Is(err, http.ErrServerClosed) {
			s.log.Warnf("unexpected (http.Server).Serve error: %v", err)
		}
		s.log.Infof("JSON-RPC listener done for %s", ln.Addr())
	}()

	// Shutdown the server on context cancellation.
	<-ctx.Done()

	s.log.Infof("JSON-RPC server shutting down...")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = s.srv.Shutdown(ctxTimeout); err != nil {
		err = fmt.Errorf("http.Server.Shutdown: %v", err)
	}

	wg.Wait()

	s.log.Infof("JSON-RPC server shutdown complete")
	return err
}

// 4 MiB request size limit
const szLimit = 1 << 22

// handlerV1 handles all https json requests. It is the http.Handler for the
// JSON-RPC service mounted on the "/rpc/v1" endpoint. The endpoint is the same
// for all methods since this is JSON-RPC with a "method" field of the JSON
// request body indicating how to process the request. Other handlers can be
// mounted on other endpoints without worry.
func (s *Server) handlerV1(w http.ResponseWriter, r *http.Request) {
	// Close the connection when response handling is completed.
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "application/json")
	r.Close = true

	bodyReader := io.LimitReader(r.Body, szLimit)
	body, err := io.ReadAll(bodyReader)
	r.Body.Close()
	if err != nil {
		http.Error(w, "error reading request body", http.StatusBadRequest)
		return
	}
	req := new(jsonrpc.Request)
	err = json.Unmarshal(body, req)
	if err != nil {
		resp := jsonrpc.NewErrorResponse(-1, jsonrpc.NewError(jsonrpc.ErrorParse, "invalid request", nil))
		s.writeJSON(w, resp, http.StatusBadRequest)
		return
	}
	s.processRequest(r.Context(), w, req)
}

// processRequest handles the jsonrpc.Request with handleRequest to call the
// appropriate function for the method, creates a response message, and writes
// it to the http.ResponseWriter.
func (s *Server) processRequest(ctx context.Context, w http.ResponseWriter, req *jsonrpc.Request) {
	resp := s.handleRequest(ctx, req)
	// Some conventions dictate 200 for everything, with the Response.Error
	// being the only sign of issue. However, a certain set of errors warrant an
	// http status code.
	statusCode := http.StatusOK
	if resp.Error != nil {
		switch resp.Error.Code {
		case jsonrpc.ErrorUnknownMethod: // other "not found" is not a 404 since the method at least existed
			statusCode = http.StatusNotFound // 404
		case jsonrpc.ErrorInvalidParams, jsonrpc.ErrorInvalidRequest, jsonrpc.ErrorParse:
			statusCode = http.StatusBadRequest // 400
		case jsonrpc.ErrorInternal:
			statusCode = http.StatusInternalServerError // 500
		}
	}
	s.writeJSON(w, resp, statusCode)
}

// writeJSONWithStatus marshals the provided interface and writes the bytes to
// the ResponseWriter with the specified response code.
func (s *Server) writeJSON(w http.ResponseWriter, thing any, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.Marshal(thing)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.log.Errorf("JSON encode error: %v", err)
		return
	}
	w.WriteHeader(code)
	_, err = w.Write(b)
	if err != nil {
		s.log.Errorf("Write error: %v", err)
	}
}

// type methodHandler func(ctx context.Context, s *Server, params json.RawMessage) (any, *jsonrpc.Error)

// routes maps routes to a handler function.
//
// TODO: instead of a map of handlers, create instance of registered request
// type and call the TxSvc method with it. But I hate reflect.
// var routes = map[string]methodHandler{
// 	jsonrpc.MethodAccount:   handleAccount,
// 	jsonrpc.MethodChainInfo: nil,
// }

// func registerMethod[T any](method string, req *T, handler func(context.Context, *T) (any, *jsonrpc.Error)) {}

// handleRequest sends the request to the correct handler function if able.
func (s *Server) handleRequest(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	if req.JSONRPC != "2.0" || req.ID == 0 {
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorInvalidRequest, "invalid json-rpc request object", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}
	if req.Method == "" {
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorUnknownMethod, "no route was supplied", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}

	// Find the correct handler for this route. xxx now using a switch in handleMethod
	// h, exists := routes[req.Method]
	// if !exists {
	// 	resp.Error = jsonrpc.NewError(jsonrpc.ErrorUnknownMethod, "unknown route", nil)
	// 	return resp
	// }

	s.log.Debug("handling request", log.String("method", req.Method))

	// call the method with the params
	result, rpcErr := handleMethod(ctx, s, jsonrpc.Method(req.Method), req.Params)
	if rpcErr != nil {
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}

	resp, err := jsonrpc.NewResponse(req.ID, result)
	if err != nil { // failed to marshal result
		s.log.Error("failed to marshal result", log.String("method", req.Method), log.Error(err))
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to encode result", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}
	return resp
}
