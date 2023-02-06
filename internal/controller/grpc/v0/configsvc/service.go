package configsvc

import (
	pb "kwil/api/protobuf/kwil/config/v0/gen/go"
	"kwil/pkg/log"
)

type Service struct {
	pb.UnimplementedConfigServiceServer
	log log.Logger
}

func NewService(logger log.Logger) *Service {
	return &Service{
		log: logger.Named("configsvc"),
	}
}
