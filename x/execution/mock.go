package execution

import (
	"context"
	"kwil/x/schema"

	"github.com/google/uuid"
)

type mockMdService struct {
}

func newMockMdService() *mockMdService {
	return &mockMdService{}
}

func (m *mockMdService) Apply(ctx context.Context, planID uuid.UUID) error {
	return nil
}

func (m *mockMdService) Plan(ctx context.Context, req schema.PlanRequest) (schema.Plan, error) {
	return schema.Plan{}, nil
}

func (m *mockMdService) GetMetadata(ctx context.Context, req schema.RequestMetadata) (schema.Metadata, error) {
	return schema.Metadata{}, nil
}
