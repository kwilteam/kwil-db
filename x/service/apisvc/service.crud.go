package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) Cud(ctx context.Context, req *apipb.CUDRequest) (*apipb.CUDResponse, error) {
	panic("not implemented")
	return &apipb.CUDResponse{}, nil
}
