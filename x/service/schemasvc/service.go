package schemasvc

import (
	"context"
	"kwil/x/proto/schemapb"
	"kwil/x/schema"

	"github.com/google/uuid"
)

type Service struct {
	schemapb.UnimplementedSchemaServiceServer

	svc schema.Service
}

func NewService(service schema.Service) *Service {
	return &Service{svc: service}
}

func (s *Service) PlanSchema(ctx context.Context, req *schemapb.PlanSchemaRequest) (*schemapb.PlanSchemaResponse, error) {
	planReq := schema.PlanRequest{
		Wallet:     req.Wallet,
		Database:   req.Database,
		SchemaData: req.Schema,
	}
	plan, err := s.svc.Plan(ctx, planReq)
	if err != nil {
		return nil, err
	}

	changes := make([]*schemapb.Change, len(plan.Changes))
	for i, change := range plan.Changes {
		changes[i] = &schemapb.Change{
			Cmd:     change.Cmd,
			Comment: change.Comment,
			Reverse: change.Reverse,
		}
	}

	return &schemapb.PlanSchemaResponse{
		Plan: &schemapb.Plan{
			PlanId:  plan.ID.String(),
			Version: plan.Version,
			Name:    plan.Name,
			Changes: changes,
		},
	}, nil

}

func (s *Service) ApplySchema(ctx context.Context, req *schemapb.ApplySchemaRequest) (*schemapb.ApplySchemaResponse, error) {
	id, err := uuid.Parse(req.PlanId)
	if err != nil {
		return nil, err
	}

	err = s.svc.Apply(ctx, id)
	return &schemapb.ApplySchemaResponse{}, err
}

func (s *Service) GetMetadata(ctx context.Context, req *schemapb.GetMetadataRequest) (*schemapb.GetMetadataResponse, error) {
	meta, err := s.svc.GetMetadata(ctx, schema.RequestMetadata{Wallet: req.Wallet, Database: req.Database})
	if err != nil {
		return nil, err
	}

	return &schemapb.GetMetadataResponse{
		Metadata: convertMetadata(meta),
	}, nil
}

func convertMetadata(meta schema.Metadata) *schemapb.Metadata {
	tables := make([]*schemapb.Table, len(meta.Tables))
	for i, table := range meta.Tables {
		tables[i] = convertTable(table)
	}
	queries := make([]*schemapb.Query, len(meta.Queries))
	for i, query := range meta.Queries {
		queries[i] = convertQuery(query)
	}

	roles := make([]*schemapb.Role, len(meta.Roles))
	for i, role := range meta.Roles {
		roles[i] = convertRole(role)
	}

	return &schemapb.Metadata{
		Name:   meta.DbName,
		Tables: tables,
	}
}

func convertTable(table schema.Table) *schemapb.Table {
	columns := make([]*schemapb.Column, len(table.Columns))
	for i, column := range table.Columns {
		columns[i] = convertColumn(column)
	}

	return &schemapb.Table{
		Name:    table.Name,
		Columns: columns,
	}
}

func convertColumn(column schema.Column) *schemapb.Column {
	return &schemapb.Column{
		Name:     column.Name,
		Type:     column.Type,
		Nullable: column.Nullable,
	}
}

func convertQuery(query schema.Query) *schemapb.Query {
	return &schemapb.Query{
		Name:      query.Name,
		Statement: query.Statement,
	}
}

func convertRole(role schema.Role) *schemapb.Role {
	return &schemapb.Role{
		Name:    role.Name,
		Queries: role.Queries,
	}
}
