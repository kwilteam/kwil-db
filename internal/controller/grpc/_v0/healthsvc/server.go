package healthsvc

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/pkg/healthcheck"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Server struct {
	grpc_health_v1.UnimplementedHealthServer
	ck healthcheck.Checker
}

func NewServer(ck healthcheck.Checker) *Server {
	ck.Start()
	return &Server{ck: ck}
}

func (s *Server) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	// here we just return overall status
	res := s.ck.Check(ctx)
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_ServingStatus(grpc_health_v1.HealthCheckResponse_ServingStatus_value[res.Status]),
	}, nil
}

func (s *Server) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return nil
}
