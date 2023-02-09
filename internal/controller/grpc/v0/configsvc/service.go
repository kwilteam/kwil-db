package configsvc

import (
	configpb "kwil/api/protobuf/config/v0"
	"kwil/pkg/log"
)

type Service struct {
	configpb.UnimplementedConfigServiceServer
	log log.Logger
}

func NewService(logger log.Logger) *Service {
	return &Service{
		log: logger.Named("configsvc"),
	}
}
