package health

import (
	"context"
	"kwil/x/healthcheck"
	"kwil/x/proto/healthpb"
)

type server struct {
	healthpb.UnimplementedHealthServer
	ck healthcheck.Checker
}

func NewServer(ck healthcheck.Checker) *server {
	ck.Start()
	return &server{ck: ck}
}

func (s *server) Check(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	// here we just return overall status
	res := s.ck.Check(ctx)
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_ServingStatus(healthpb.HealthCheckResponse_ServingStatus_value[res.Status]),
	}, nil
}
