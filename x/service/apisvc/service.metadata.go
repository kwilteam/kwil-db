package apisvc

import (
	"context"

	"kwil/x/proto/apipb"
)

func (s *Service) GetMetadata(ctx context.Context, req *apipb.GetMetadataRequest) (*apipb.GetMetadataResponse, error) {

	/*
		md, err := s.md.GetDatabase(ctx, req.Owner, req.Name)
		if err != nil {
			return nil, err
		}

		// marshal the metadata into a proto message
		bts, err := json.Marshal(md)
		if err != nil {
			return nil, err
		}

		var m apipb.GetMetadataResponse
		err = json.Unmarshal(bts, &m)
		if err != nil {
			return nil, err
		}*/
	mdr := &apipb.GetMetadataResponse{
		Name: "test",
		Queries: []*apipb.Query{
			{
				Name:      "query2",
				Statement: "insert...",
				Inputs: []*apipb.Input{
					{
						Name: "input1",
						Type: "string",
					},
					{
						Name: "input2",
						Type: "string",
					},
				},
				Outputs: []*apipb.Output{
					{
						Name: "output1",
						Type: "string",
					},
				},
			},
			{
				Name:      "query2",
				Statement: "insert ...",
				Inputs: []*apipb.Input{
					{
						Name: "input1",
						Type: "string",
					},
				},
				Outputs: []*apipb.Output{
					{
						Name: "output1",
						Type: "string",
					},
				},
			},
		},
		Roles: []*apipb.Role{
			{
				Name:    "test",
				Queries: []string{"query1", "query2"},
			},
		},
		Tables: []*apipb.Table{
			{
				Name: "table1",
				Columns: []*apipb.Column{
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
