package server

import (
	"net"

	"github.com/kwilteam/kwil-db/core/log"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	server   *grpc.Server
	logger   log.Logger
	listener net.Listener

	srvOpts []grpc.ServerOption
}

func New(logger log.Logger, lis net.Listener, opts ...Option) *Server {
	l := *logger.WithOptions(zap.WithCaller(false))

	recoveryFunc := func(p interface{}) error {
		l.Error("panic triggered", zap.Any("panic", p))
		return status.Errorf(codes.Unknown, "unknown error")
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryFunc),
	}

	s := &Server{
		logger:   l,
		listener: lis,
	}

	for _, opt := range opts {
		opt(s)
	}
	srvOpts := append(s.srvOpts, grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		SimpleInterceptorLogger(&l),
	))

	s.server = grpc.NewServer(srvOpts...)

	return s
}

func (s *Server) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
}

func (s *Server) Start() error {
	return s.server.Serve(s.listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
