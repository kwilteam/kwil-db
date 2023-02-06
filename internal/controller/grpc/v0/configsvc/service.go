package configsvc

import (
	pb "kwil/api/protobuf/config/v0/gen/go"
	"kwil/pkg/logger"
)

type Service struct {
	pb.UnimplementedConfigServiceServer
	log logger.Logger
}

func NewService(logger logger.Logger) *Service {
	return &Service{
		log: logger.Named("configsvc"),
	}
}
