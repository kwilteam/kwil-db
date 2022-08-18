package rest

import (
	"context"
)

type Write struct {
	ID string `json:"id"`
}

type Store interface {
	GetWrite(context.Context, string) (Write, error)
}

// Struct for service logic
type Service struct {
	Store Store
}

// NewService returns a pointer Service.
func NewService() *Service {
	return &Service{}
}

func (s *Service) Write(ctx context.Context, w Write) (Write, error) {
	return w, nil
}
