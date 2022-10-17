package request

import "context"

const MANAGER_ALIAS = "service.REQUEST_MANAGER" // todo: ensure unique

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
