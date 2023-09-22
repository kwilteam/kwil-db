package admin

import (
	"context"

	"github.com/kwilteam/kwil-db/internal/pkg/version"
	"github.com/kwilteam/kwil-db/pkg/log"

	admpb "github.com/kwilteam/kwil-db/api/protobuf/admin/v0"
)

type AdminSvcOpt func(*Service)

func WithLogger(logger log.Logger) AdminSvcOpt {
	return func(s *Service) {
		s.log = logger
	}
}

// Service is the implementation of the admpb.AdminServiceServer methods.
type Service struct {
	admpb.UnimplementedAdminServiceServer

	log log.Logger
}

// NewService constructs a new Service.
func NewService(opts ...AdminSvcOpt) *Service {
	s := &Service{
		log: log.NewNoOp(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Ping responds to any ping request with "pong".
func (svc *Service) Ping(ctx context.Context, req *admpb.PingRequest) (*admpb.PingResponse, error) {
	return &admpb.PingResponse{Message: "pong"}, nil
}

// Version reports the compile-time kwild version.
func (svc *Service) Version(ctx context.Context, req *admpb.VersionRequest) (*admpb.VersionResponse, error) {
	return &admpb.VersionResponse{
		VersionString: version.KwilVersion,
	}, nil
}

func (svc *Service) NodeName(ctx context.Context) (string, error) {
	return "dummyName" /*s.Node.Moniker()*/, nil
}
