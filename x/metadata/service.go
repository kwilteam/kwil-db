package metadata

import (
	"context"

	"ksl/ast"
	"ksl/lift"
	"ksl/sqlclient"
	"ksl/sqlmigrate"
)

type Service interface {
	Plan(context.Context, SchemaRequest) (Plan, error)
	Apply(context.Context, SchemaRequest) error
	GetMetadata(context.Context, RequestMetadata) (Metadata, error)
}

type sqlservice struct {
	connector Connector
}

func NewService(connector Connector) Service {
	return &sqlservice{
		connector: connector,
	}
}

func (s *sqlservice) planInternal(ctx context.Context, req SchemaRequest, apply bool) (sqlmigrate.MigrationPlan, error) {
	ksch := ast.Parse(req.SchemaData, "<schema>")

	if ksch.HasErrors() {
		return sqlmigrate.MigrationPlan{}, ksch.Diagnostics
	}

	url, err := s.connector.GetConnectionInfo(req.Wallet)
	if err != nil {
		return sqlmigrate.MigrationPlan{}, err
	}

	client, err := sqlclient.Open(url)
	if err != nil {
		return sqlmigrate.MigrationPlan{}, err
	}
	defer client.Close()

	source, err := client.DescribeContext(ctx, req.Database)
	if err != nil {
		return sqlmigrate.MigrationPlan{}, err
	}

	target := lift.Sql(ksch, req.Database)

	plan, err := client.PlanMigration(ctx, source, target)
	if err != nil {
		return sqlmigrate.MigrationPlan{}, err
	}

	if apply {
		err = client.ApplyMigration(ctx, plan)
	}

	return plan, err
}

func (s *sqlservice) Plan(ctx context.Context, req SchemaRequest) (Plan, error) {
	plan, err := s.planInternal(ctx, req, false)
	if err != nil {
		return Plan{}, err
	}

	return convertPlan(plan), nil
}

func (s *sqlservice) Apply(ctx context.Context, req SchemaRequest) error {
	_, err := s.planInternal(ctx, req, true)
	return err
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
