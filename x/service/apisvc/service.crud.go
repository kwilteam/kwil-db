package apisvc

import (
	"context"
	"encoding/json"
	"fmt"

	"kwil/x/execution"
	"kwil/x/proto/apipb"
)

func (s *Service) Cud(ctx context.Context, req *apipb.CUDRequest) (*apipb.CUDResponse, error) {
	_, err := s.p.GetPrice(ctx)
	if err != nil {
		return nil, err
	}
	panic("not implemented")
}

func (s *Service) Read(ctx context.Context, req *apipb.ReadRequest) (*apipb.ReadResponse, error) {
	bi, err := json.Marshal(req.Inputs)
	if err != nil {
		return nil, err
	}

	var ins []execution.Input
	err = json.Unmarshal(bi, &ins)
	if err != nil {
		return nil, err
	}

	res, err := s.e.Read(ctx, req.Owner, req.Database, req.Query, ins)
	if err != nil {
		return nil, err
	}

	fmt.Println(res)

	return nil, nil

}
