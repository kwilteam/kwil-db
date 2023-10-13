package server

import "google.golang.org/grpc"

type Option func(*Server)

func WithSrvOpt(srvOpt grpc.ServerOption) Option {
	return func(s *Server) {
		s.srvOpts = append(s.srvOpts, srvOpt)
	}
}
