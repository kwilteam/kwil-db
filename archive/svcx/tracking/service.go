package tracking

import (
	"context"
	"kwil/x/async"
	"kwil/x/cfgx"
)

const SERVICE_ALIAS = "tracking-service" // todo: ensure unique

func New(_ cfgx.Config) (Service, error) {
	panic("implement me")
}

type ID string
type Response async.Task[Item]

type Service interface {
	// Submit creates a new item to track commands submitted for execution
	Submit(ctx context.Context, source_identity string, correlation_id string) Response

	// Update used to update an item as it moves through the system
	Update(ctx context.Context, tracking_id ID, status Status) async.Action

	// Find used to get the current item status
	Find(ctx context.Context, tracking_id ID) Response

	// FindByExternalKey used to get the item by the external key
	FindByExternalKey(ctx context.Context, correlation_id string) Response
}

type service struct{}

func (m *service) Submit(ctx context.Context, source_identity string, correlation_id string) Response {
	//TODO implement me

	panic("implement me")
}

func (m *service) Update(ctx context.Context, tracking_id ID, status Status) async.Action {
	//TODO implement me
	panic("implement me")
}

func (m *service) Find(ctx context.Context, tracking_id ID) Response {
	//TODO implement me
	panic("implement me")
}

func (m *service) FindByExternalKey(ctx context.Context, correlation_id string) Response {
	//TODO implement me
	panic("implement me")
}
