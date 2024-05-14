package server

import (
	"context"
	"net"
	"time"

	"github.com/kwilteam/kwil-db/core/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// unaryTimeoutInterceptor is a unary server interceptor that sets a timeout for each request.
func unaryTimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan struct{})
		var res any
		var err error

		go func() {
			res, err = handler(ctx, req)
			close(done)
		}()

		select {
		case <-ctx.Done(): // canceled, handler should return soon, but don't wait
			return nil, status.Errorf(codes.DeadlineExceeded, "request timed out")
		case <-done: // finished before timeout
			return res, err
		}
	}
}

type Server struct {
	server   *grpc.Server
	logger   log.Logger
	listener net.Listener

	// config used only in constructor :(
	srvOpts []grpc.ServerOption
	timeout time.Duration
}

const defaultTimeout = 30 * time.Second

func New(logger log.Logger, lis net.Listener, opts ...Option) *Server {
	l := *logger.WithOptions(zap.WithCaller(false))

	recoveryFunc := func(p any) error {
		l.Error("panic triggered", zap.Any("panic", p))
		return status.Errorf(codes.Unknown, "unknown error")
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryFunc),
	}

	s := &Server{
		logger:   l,
		listener: lis,
		timeout:  defaultTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}
	srvOpts := append(s.srvOpts, grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		unaryTimeoutInterceptor(s.timeout),
		SimpleInterceptorLogger(&l),
	))

	s.server = grpc.NewServer(srvOpts...)

	return s
}

func (s *Server) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
}

func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
