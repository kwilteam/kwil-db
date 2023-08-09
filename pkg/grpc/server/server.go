package server

import (
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	server   *grpc.Server
	logger   log.Logger
	listener net.Listener
}

func New(logger log.Logger, lis net.Listener, opts ...Option) *Server {
	l := *logger.Named("grpcServer").WithOptions(zap.WithCaller(false))

	recoveryFunc := func(p interface{}) error {
		l.Error("panic triggered", zap.Any("panic", p))
		return status.Errorf(codes.Unknown, "unknown error")
	}
	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryFunc),
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall),
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(recoveryOpts...),
			logging.UnaryServerInterceptor(InterceptorLogger(&l), loggingOpts...),
		),
	)

	s := &Server{
		server:   server,
		logger:   l,
		listener: lis,
	}

	for _, opt := range opts {
		opt(s)
	}

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
