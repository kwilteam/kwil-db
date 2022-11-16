package execution

import (
	"context"
	"kwil/x/metadata"

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

func (m *mockMdService) Plan(ctx context.Context, req metadata.PlanRequest) (metadata.Plan, error) {
	return metadata.Plan{}, nil
}

func (m *mockMdService) GetMetadata(ctx context.Context, req metadata.RequestMetadata) (metadata.Metadata, error) {
	return metadata.Metadata{}, nil
}
