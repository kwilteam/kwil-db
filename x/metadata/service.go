package metadata

import (
	"context"

	"ksl/schema"
	"ksl/sqlclient"
	"ksl/sqlschema"

	"github.com/google/uuid"
)

type Service interface {
	Plan(context.Context, PlanRequest) (Plan, error)
	Apply(context.Context, uuid.UUID) error
	GetMetadata(context.Context, RequestMetadata) (Metadata, error)
}

type sqlservice struct {
	connector Connector
	planner   PlanRepository
}

func NewService(connector Connector, planner PlanRepository) Service {
	return &sqlservice{
		connector: connector,
		planner:   planner,
	}
}

func (s *sqlservice) Plan(ctx context.Context, req PlanRequest) (Plan, error) {
	ksch := schema.Parse(req.SchemaData, "<schema>")

	if ksch.HasErrors() {
		return Plan{}, ksch.Diagnostics
	}

	url, err := s.connector.GetConnectionInfo(req.Wallet)
	if err != nil {
		return Plan{}, err
	}

	client, err := sqlclient.Open(url)
	if err != nil {
		return Plan{}, err
	}
	defer client.Close()

	source, err := client.DescribeContext(ctx, req.Database)
	if err != nil {
		return Plan{}, err
	}

	target := sqlschema.CalculateSqlSchema(ksch, req.Database)

	plan, err := client.PlanMigration(ctx, source, target)
	if err != nil {
		return Plan{}, err
	}

	planID, err := s.planner.SavePlan(req.Wallet, req.Database, req.SchemaData)
	if err != nil {
		return Plan{}, err
	}

	return convertPlan(planID, plan), nil
}

func (s *sqlservice) Apply(ctx context.Context, planID uuid.UUID) error {
	info, err := s.planner.GetPlanInfo(planID)
	if err != nil {
		return err
	}

	ksch := schema.Parse(info.Data, "<schema>")

	if ksch.HasErrors() {
		return ksch.Diagnostics
	}

	target := sqlschema.CalculateSqlSchema(ksch, info.Database)

	url, err := s.connector.GetConnectionInfo(info.Wallet)
	if err != nil {
		return err
	}

	client, err := sqlclient.Open(url)
	if err != nil {
		return err
	}
	defer client.Close()

	source, err := client.DescribeContext(ctx, info.Database)
	if err != nil {
		return err
	}

	plan, err := client.PlanMigration(ctx, source, target)
	if err != nil {
		return err
	}

	return client.ApplyMigration(ctx, plan)
}

func (s *sqlservice) GetMetadata(ctx context.Context, req RequestMetadata) (Metadata, error) {
	url, err := s.connector.GetConnectionInfo(req.Wallet)
	if err != nil {
		return Metadata{}, err
	}

	client, err := sqlclient.Open(url)
	if err != nil {
		return Metadata{}, err
	}
	defer client.Close()

	source, err := client.DescribeContext(ctx, req.Database)
	if err != nil {
		return Metadata{}, err
	}

	return convertDatabase(source), nil
}
