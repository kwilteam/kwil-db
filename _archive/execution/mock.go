package execution

import (
	"context"
	"kwil/_archive/metadata"
)

type mockMdService struct {
}

func newMockMdService() *mockMdService {
	return &mockMdService{}
}

func (m *mockMdService) Apply(ctx context.Context, req metadata.SchemaRequest) error {
	return nil
}

func (m *mockMdService) Plan(ctx context.Context, req metadata.SchemaRequest) (metadata.Plan, error) {
	return metadata.Plan{}, nil
}

func (m *mockMdService) GetMetadata(ctx context.Context, req metadata.RequestMetadata) (metadata.Metadata, error) {
	mdr := metadata.Metadata{
		DbName: "test",
		Queries: []metadata.Query{
			{
				Name: "query1",
				Inputs: []metadata.Param{
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
				Name: "query2",
				Inputs: []metadata.Param{
					{
						Name: "input1",
						Type: "string",
					},
				},
			},
		},
		Roles: []metadata.Role{
			{
				Name:    "test",
				Queries: []string{"query1", "query2"},
			},
		},
		Tables: []metadata.Table{
			{
				Name: "table1",
				Columns: []metadata.Column{
					{
						Name: "column1",
						Type: "string",
					},
					{
						Name: "column2",
						Type: "string",
					},
				},
			},
		},
	}
	return mdr, nil
}
