package server

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	server *grpc.Server
	logger log.Logger
}

func New(logger log.Logger, opts ...Option) *Server {
	l := *logger.Named("grpcServer").WithOptions(zap.WithCaller(false))

	recoveryFunc := func(p interface{}) error {
		l.Error("grpc: panic serving", zap.Any("panic", p))
		return nil
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
		server: server,
		logger: l,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
}

func (s *Server) Serve(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
