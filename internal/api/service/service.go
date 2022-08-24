package service

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/types"
)

type KDB interface {
	CreateDatabase(ctx context.Context, db *types.CreateDatabase) error
}

// Service Struct for service logic
type Service struct {
	KDB KDB
}

// NewService returns a pointer Service.
func NewService() *Service {
	return &Service{}
}
