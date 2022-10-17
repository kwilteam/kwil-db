package request

import (
	"context"
	"sync"
)

const MANAGER_ALIAS = "service.REQUEST_MANAGER" // todo: ensure unique

var managerFn func() Manager

func init() {
	once := sync.Once{}
	var m Manager
	managerFn = func() Manager {
		once.Do(func() {
			m = &manager{}
		})
		return m
	}
}

func GetManager() Manager {
	return managerFn()
}

type Manager interface {
	// Create creates a new request to track commands submitted for execution
	Create(ctx context.Context, source_identifier string) (Info, error)

	// Update used to update requests as it moves through the system
	Update(ctx context.Context, status Status) error

	// FindByRequestID used to get the current request status
	FindByRequestID(ctx context.Context, request_id string) (Info, error)

	// Find used to get the current request by the source identifying tx key
	Find(ctx context.Context, idempotent_key string) (Info, error)
}

type manager struct{}

func (m *manager) Create(ctx context.Context, source_identifier string) (Info, error) {
	//TODO implement me

	panic("implement me")
}

func (m *manager) Update(ctx context.Context, status Status) error {
	//TODO implement me
	panic("implement me")
}

func (m *manager) FindByRequestID(ctx context.Context, request_id string) (Info, error) {
	//TODO implement me
	panic("implement me")
}

func (m *manager) Find(ctx context.Context, idempotent_key string) (Info, error) {
	//TODO implement me
	panic("implement me")
}
