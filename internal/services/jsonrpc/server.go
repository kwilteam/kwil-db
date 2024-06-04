package rpcserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"
	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/openrpc"
)

// The endpoint path is constant for now.
const (
	pathRPCV1  = "/rpc/v1"
	pathSpecV1 = "/spec/v1"
)

// Server is a JSON-RPC server.
type Server struct {
	srv            *http.Server
	unix           bool // the listener's network should be "unix" instead of "tcp"
	log            log.Logger
	methodHandlers map[jsonrpc.Method]MethodHandler
	methodDefs     map[string]*openrpc.MethodDefinition
	specInfo       *openrpc.Info
	spec           json.RawMessage
	authSHA        []byte
	tlsCfg         *tls.Config
}

type serverConfig struct {
	pass       string
	tlsConfig  *tls.Config
	timeout    time.Duration
	enableCORS bool
	specInfo   *openrpc.Info
}

type Opt func(*serverConfig)

// WithPass will require a password in the request header's Authorization value
// in "Basic" formatting. Don't use this without TLS; either terminate TLS in an
// upstream reverse proxy, or use WithTLS with Certificates provided.
func WithPass(pass string) Opt {
	return func(c *serverConfig) {
		c.pass = pass
	}
}

// WithTLS provides a tls.Config to use with tls.NewListener around the standard
// net.Listener.
func WithTLS(cfg *tls.Config) Opt {
	return func(c *serverConfig) {
		c.tlsConfig = cfg
	}
}

// WithServerInfo sets the OpenRPC "info" section to use when serving the
// OpenRPC JSON specification either via a spec REST endpoint or the
// rpc.discover JSON-RPC method.
func WithServerInfo(specInfo *openrpc.Info) Opt {
	return func(c *serverConfig) {
		c.specInfo = specInfo
	}
}

// WithTimeout specifies a timeout on all RPC requests that when exceeded will
// cancel the request.
func WithTimeout(timeout time.Duration) Opt {
	return func(c *serverConfig) {
		c.timeout = timeout
	}
}

func WithCORS() Opt {
	return func(c *serverConfig) {
		c.enableCORS = true
	}
}

// checkAddr cleans the address, and indicates if it is a unix socket (local
// filesystem path). The addr for NewServer should be a host:port style string,
// but if it is a URL, this will attempt to get the host and port from it.
func checkAddr(addr string) (string, bool, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = addr
			port = "8485"
		} else if strings.Contains(err.Error(), "too many colons in address") {
			u, err := url.Parse(addr)
			if err != nil {
				return "", false, err
			}
			host, port = u.Host, u.Port()
		} else {
			return "", false, err
		}
	}

	if strings.HasPrefix(host, "/") { // unix, no port
		return host, true, nil
	}

	return net.JoinHostPort(host, port), false, nil
}

// defaultWriteTimeout is the default WriteTimeout for the http.Server.
const defaultWriteTimeout = 45 * time.Second

var (
	defaultSpecInfo = &openrpc.Info{
		Title:       "Kwil DB RPC service",
		Description: `The JSON-RPC service for Kwil DB.`,
		License: &openrpc.License{ // the spec file's license
			Name: "CC0-1.0",
			URL:  "https://creativecommons.org/publicdomain/zero/1.0/legalcode",
		},
		Version: "0.1.0",
	}
)

// NewServer creates a new JSON-RPC server. Use RegisterMethodHandler or
// RegisterSvc to add method handlers.
func NewServer(addr string, log log.Logger, opts ...Opt) (*Server, error) {
	addr, isUNIX, err := checkAddr(addr)
	if err != nil {
		return nil, err
	}
	if isUNIX { // prepare for the socket file
		err = os.MkdirAll(filepath.Dir(addr), 0700) // ensure parent dir exists
		if err != nil {
			return nil, fmt.Errorf("failed to create admin service unix socket directory at %q: %w",
				filepath.Dir(addr), err)
		}

		err = syscall.Unlink(addr) // if left from last run
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to build grpc server: %w", err)
		}
	}

	cfg := &serverConfig{
		timeout:  defaultWriteTimeout,
		specInfo: defaultSpecInfo,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	mux := http.NewServeMux() // http.DefaultServeMux has the pprof endpoints mounted

	disconnectTimeout := cfg.timeout + 5*time.Second // for jsonRPCTimeoutHandler to respond, don't disconnect immediately
	srv := &http.Server{
		Addr:              addr, // only used with srv.ListenAndServe, not Serve
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,  // receiving request body should not take longer
		WriteTimeout:      disconnectTimeout, // full request handling: receive request, handle request, AND send response
	}

	if srv.ReadTimeout > srv.WriteTimeout {
		srv.ReadTimeout = srv.WriteTimeout
	}
	if srv.ReadHeaderTimeout > srv.ReadTimeout {
		srv.ReadHeaderTimeout = srv.ReadTimeout
	}

	s := &Server{
		srv:            srv,
		unix:           isUNIX,
		log:            log,
		methodHandlers: make(map[jsonrpc.Method]MethodHandler),
		methodDefs:     make(map[string]*openrpc.MethodDefinition),
		specInfo:       cfg.specInfo,
		tlsCfg:         cfg.tlsConfig,
	}

	if cfg.pass != "" {
		authSHA := sha256.Sum256([]byte(cfg.pass))
		s.authSHA = slices.Clone(authSHA[:])
	} // otherwise no basic auth check

	var h http.Handler
	h = http.HandlerFunc(s.handlerV1)
	h = http.MaxBytesHandler(h, 1<<22)
	// amazingly, exceeding the server's write timeout does not cancel request
	// contexts: https://github.com/golang/go/issues/59602
	// So, we add a timeout to the Request's context.
	h = jsonRPCTimeoutHandler(h, cfg.timeout, log)
	if cfg.enableCORS {
		h = corsHandler(h)
	}
	h = recoverer(h, log)

	mux.Handle(pathRPCV1, h)

	var specHandler http.Handler
	specHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("content-type", "application/json; charset=utf-8")
		http.ServeContent(w, r, "openrpc.json", time.Time{}, bytes.NewReader(s.spec))
	})
	specHandler = corsHandler(specHandler)
	mux.Handle(pathSpecV1, specHandler)

	return s, nil
}

// jsonRPCTimeoutHandler runs the handler with a time limit. This middleware
// also logs the total request duration since it is a logical place that deals
// with request timing. This duration is the total elapsed duration that
// includes reading the full request, unmarshalling, running the corresponding
// method handler, marshalling the response, and transmitting the entire
// response to the client. This duration is at least as long as the time logged
// in processRequest, which pertains only to the handling of the request and
// thus reflects the server's computational burden, while this duration provides
// insight into the latencies introduced by bandwidth and marshalling.
func jsonRPCTimeoutHandler(h http.Handler, timeout time.Duration, logger log.Logger) http.Handler {
	// We'll respond with a jsonrpc.Response type, but the request handler is
	// downstream and we don't have the request ID.
	resp := jsonrpc.NewErrorResponse(-1, jsonrpc.NewError(jsonrpc.ErrorTimeout, "RPC timeout", nil))
	respMsg, _ := json.Marshal(resp)
	h = http.TimeoutHandler(h, timeout, string(respMsg))

	// Log total request handling time (including transfer).
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now().UTC()
		defer func() {
			logger.Debug("request handling complete", log.Duration("total_elapsed", time.Since(t0)))
		}()
		// NOTE, to give downstream handlers access to t0 instead of a defer here:
		// ctx := context.WithValue(r.Context(), CtxStartTime, t0); r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

// corsHandler adds CORS headers to the response. We don't need sophisticated
// cors handling here (not really kwild's concern, there should be other services
// like LBs or KGW do that), so we just allow them.
// NOTE: if this server is served behind KGW, those headers will be stripped.
func corsHandler(h http.Handler) http.Handler {
	allowMethods := "GET, POST, OPTIONS"
	allowHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, ResponseType, Range"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", allowMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Preflight request
		if r.Method == http.MethodOptions {
			return
		}

		// Other SIMPLE requests and non-cors requests
		h.ServeHTTP(w, r)
	})
}

func recoverer(h http.Handler, log log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(rvr)
				}

				debugStack := debug.Stack()
				log.Errorf("panic:\n%v", string(debugStack))

				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func (s *Server) Serve(ctx context.Context) error {
	network := "tcp"
	if s.unix {
		network = "unix"
	}
	ln, err := net.Listen(network, s.srv.Addr)
	if err != nil {
		return err
	}
	if s.tlsCfg != nil {
		ln = tls.NewListener(ln, s.tlsCfg)
	}
	if s.unix {
		if err = os.Chmod(s.srv.Addr, 0755); err != nil {
			ln.Close()
			return err
		}
	}
	s.log.Info("JSON-RPC server listening", log.String("address", ln.Addr().String()))
	return s.ServeOn(ctx, ln)
}

func openRPCSpec(specInfo *openrpc.Info, methodDefs map[string]*openrpc.MethodDefinition) *openrpc.Spec {
	knownSchemas := make(map[reflect.Type]openrpc.Schema)
	methods := openrpc.InventoryAPI(methodDefs, knownSchemas)
	schemas := make(map[string]openrpc.Schema)
	for _, schema := range knownSchemas {
		schemas[schema.Name()] = schema
	}
	return &openrpc.Spec{
		OpenRPC: "1.2.4",
		Info:    *specInfo,
		Methods: methods,
		Components: openrpc.Components{
			Schemas: schemas,
		},
	}
}

func (s *Server) ServeOn(ctx context.Context, ln net.Listener) error {
	// With all services registered, only now can we generate the RPC spec.
	spec := openRPCSpec(s.specInfo, s.methodDefs)
	var err error
	s.spec, err = json.Marshal(spec)
	if err != nil {
		return err
	}

	s.RegisterMethodHandler(
		"rpc.discover",
		MakeMethodHandler(func(context.Context, *any) (*json.RawMessage, *jsonrpc.Error) {
			return &s.spec, nil
		}),
	)

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
	err = s.srv.Shutdown(ctxTimeout)
	if err != nil {
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

	if s.authSHA != nil {
		// r.SetBasicAuth("", "passwords")
		_, pass, haveAuth := r.Response.Request.BasicAuth() // r.Header.Get("Authorization")
		if !haveAuth {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Reveal nothing about the configured pass in verification time.
		authSHA := sha256.Sum256([]byte(pass))
		if subtle.ConstantTimeCompare(s.authSHA, authSHA[:]) != 1 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	}

	bodyReader := http.MaxBytesReader(w, r.Body, szLimit)
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
	// Handle and time the request.
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

	// Write the response
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

func zeroID(id any) bool {
	if id == nil {
		return true // would be !rv.IsValid()
	}
	rv := reflect.ValueOf(id)
	if rv.IsZero() {
		return true
	}
	rt := reflect.TypeOf(id) // already caught nil interface and nil ptr
	if rt.Kind() == reflect.Ptr {
		return zeroID(rv.Elem().Interface())
	}
	return false // already did rv.IsZero
}

// handleRequest sends the request to the correct handler function if able.
func (s *Server) handleRequest(ctx context.Context, req *jsonrpc.Request) *jsonrpc.Response {
	if req.JSONRPC != "2.0" || zeroID(req.ID) {
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorInvalidRequest, "invalid json-rpc request object", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}
	if req.Method == "" {
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorUnknownMethod, "no route was supplied", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}

	s.log.Debug("handling request", log.String("method", req.Method))
	t0 := time.Now().UTC() // time only the handling (pertains to server utilization)

	// call the method with the params
	result, rpcErr := s.handleMethod(ctx, jsonrpc.Method(req.Method), req.Params)
	if rpcErr != nil {
		s.log.Info("request failure", log.String("method", req.Method),
			log.Duration("elapsed", time.Since(t0)), log.Int("code", rpcErr.Code),
			log.String("message", rpcErr.Message))

		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}

	s.log.Info("request success", log.String("method", req.Method),
		log.Duration("elapsed", time.Since(t0)))

	resp, err := jsonrpc.NewResponse(req.ID, result)
	if err != nil { // failed to marshal result
		s.log.Error("failed to marshal result", log.String("method", req.Method), log.Error(err))
		rpcErr := jsonrpc.NewError(jsonrpc.ErrorResultEncoding, "failed to encode result", nil)
		return jsonrpc.NewErrorResponse(req.ID, rpcErr)
	}
	return resp
}
