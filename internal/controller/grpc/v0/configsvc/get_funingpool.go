package configsvc

import (
	"context"
	pb "kwil/api/protobuf/kwil/config/v0/gen/go"
)

func (s *Service) GetFundingPool(context.Context, *pb.GetFundingPoolRequest) (*pb.GetFundingPoolResponse, error) {
	panic("implement me")
	return nil, nil
}
