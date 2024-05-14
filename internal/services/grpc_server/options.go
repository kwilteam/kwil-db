package server

import (
	"time"

	"google.golang.org/grpc"
)

type Option func(*Server)

func WithSrvOpt(srvOpt grpc.ServerOption) Option {
	return func(s *Server) {
		s.srvOpts = append(s.srvOpts, srvOpt)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.timeout = timeout
	}
}
