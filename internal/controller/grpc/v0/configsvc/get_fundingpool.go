package configsvc

import (
	"context"
	configpb "kwil/api/protobuf/config/v0"
)

func (s *Service) GetFundingPool(context.Context, *configpb.GetFundingPoolRequest) (*configpb.GetFundingPoolResponse, error) {
	panic("implement me")
	//return nil, nil
}
