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
	mdr := schema.Metadata{
		DbName: "test",
		Queries: []schema.Query{
			{
				Name:      "query1",
				Statement: "insert...",
				Inputs: []schema.Input{
					{
						Name: "input1",
						Type: "string",
					},
					{
						Name: "input2",
						Type: "string",
					},
				},
			},
			{
				Name:      "query2",
				Statement: "insert ...",
				Inputs: []schema.Input{
					{
						Name: "input1",
						Type: "string",
					},
				},
			},
		},
		Roles: []schema.Role{
			{
				Name:    "test",
				Queries: []string{"query1", "query2"},
			},
		},
		Tables: []schema.Table{
			{
				Name: "table1",
				Columns: []schema.Column{
					{
						Name:     "column1",
						Type:     "string",
						Nullable: true,
					},
					{
						Name:     "column2",
						Type:     "string",
						Nullable: false,
					},
				},
			},
		},
	}
	return mdr, nil
}
