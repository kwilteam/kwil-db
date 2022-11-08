package apisvc

import (
	"context"
	"kwil/x/proto/apipb"
	"kwil/x/schema"

	"github.com/google/uuid"
)

func (s *Service) PlanSchema(ctx context.Context, req *apipb.PlanSchemaRequest) (*apipb.PlanSchemaResponse, error) {
	planReq := schema.PlanRequest{
		Wallet:     req.Wallet,
		Database:   req.Database,
		SchemaData: req.Schema,
	}
	plan, err := s.md.Plan(ctx, planReq)
	if err != nil {
		return nil, err
	}

	changes := make([]*apipb.Change, len(plan.Changes))
	for i, change := range plan.Changes {
		changes[i] = &apipb.Change{
			Cmd:     change.Cmd,
			Comment: change.Comment,
			Reverse: change.Reverse,
		}
	}

	return &apipb.PlanSchemaResponse{
		Plan: &apipb.Plan{
			PlanId:  plan.ID.String(),
			Version: plan.Version,
			Name:    plan.Name,
			Changes: changes,
		},
	}, nil

}

func (s *Service) ApplySchema(ctx context.Context, req *apipb.ApplySchemaRequest) (*apipb.ApplySchemaResponse, error) {
	id, err := uuid.Parse(req.PlanId)
	if err != nil {
		return nil, err
	}

	err = s.md.Apply(ctx, id)
	return &apipb.ApplySchemaResponse{}, err
}

func (s *Service) GetMetadata(ctx context.Context, req *apipb.GetMetadataRequest) (*apipb.GetMetadataResponse, error) {
	meta, err := s.md.GetMetadata(ctx, schema.RequestMetadata{Wallet: req.Wallet, Database: req.Database})
	if err != nil {
		return nil, err
	}

	return &apipb.GetMetadataResponse{
		Metadata: convertMetadata(meta),
	}, nil
}

func convertMetadata(meta schema.Metadata) *apipb.Metadata {
	tables := make([]*apipb.Table, len(meta.Tables))
	for i, table := range meta.Tables {
		tables[i] = convertTable(table)
	}
	queries := make([]*apipb.Query, len(meta.Queries))
	for i, query := range meta.Queries {
		queries[i] = convertQuery(query)
	}

	roles := make([]*apipb.Role, len(meta.Roles))
	for i, role := range meta.Roles {
		roles[i] = convertRole(role)
	}

	return &apipb.Metadata{
		Name:        meta.DbName,
		Tables:      tables,
		Queries:     queries,
		Roles:       roles,
		DefaultRole: meta.DefaultRole,
	}
}

func convertTable(table schema.Table) *apipb.Table {
	columns := make([]*apipb.Column, len(table.Columns))
	for i, column := range table.Columns {
		columns[i] = convertColumn(column)
	}

	return &apipb.Table{
		Name:    table.Name,
		Columns: columns,
	}
}

func convertColumn(column schema.Column) *apipb.Column {
	return &apipb.Column{
		Name:     column.Name,
		Type:     column.Type,
		Nullable: column.Nullable,
	}
}

func convertQuery(query schema.Query) *apipb.Query {
	return &apipb.Query{
		Name:      query.Name,
		Statement: query.Statement,
	}
}

func convertRole(role schema.Role) *apipb.Role {
	return &apipb.Role{
		Name:    role.Name,
		Queries: role.Queries,
	}
}
