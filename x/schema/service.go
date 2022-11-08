package schema

import (
	"context"

	"github.com/google/uuid"
	"ksl/kslparse"
	"ksl/sqlclient"
	"ksl/sqlspec"
)

type Service interface {
	Plan(context.Context, PlanRequest) (Plan, error)
	Apply(context.Context, uuid.UUID) error
	GetDatabase(context.Context, string, string) (Database, error)
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
	parser := kslparse.NewParser()
	_, diags := parser.Parse(req.SchemaData, "<schema>")

	if diags.HasErrors() {
		return Plan{}, diags
	}

	target, diags := sqlspec.Decode(parser.FileSet())
	if diags.HasErrors() {
		return Plan{}, diags
	}

	url, err := s.connector.GetConnectionInfo(req.Wallet)
	if err != nil {
		return Plan{}, err
	}

	client, err := sqlclient.Open(ctx, url)
	if err != nil {
		return Plan{}, err
	}
	defer client.Close()

	targetOpts := &sqlspec.InspectRealmOption{Schemas: []string{req.Database}}
	source, err := client.InspectRealm(ctx, targetOpts)
	if err != nil {
		return Plan{}, err
	}

	changes, err := client.RealmDiff(source, target)
	if err != nil {
		return Plan{}, err
	}

	plan, err := client.PlanChanges(changes)
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

	parser := kslparse.NewParser()
	_, diags := parser.Parse(info.Data, "<schema>")

	if diags.HasErrors() {
		return diags
	}

	target, diags := sqlspec.Decode(parser.FileSet())
	if diags.HasErrors() {
		return diags
	}

	url, err := s.connector.GetConnectionInfo(info.Wallet)
	if err != nil {
		return err
	}

	client, err := sqlclient.Open(ctx, url)
	if err != nil {
		return err
	}
	defer client.Close()

	opts := &sqlspec.InspectRealmOption{Schemas: []string{info.Database}}
	source, err := client.InspectRealm(ctx, opts)
	if err != nil {
		return err
	}

	changes, err := client.RealmDiff(source, target)
	if err != nil {
		return err
	}

	return client.ApplyChanges(ctx, changes)
}

func (s *sqlservice) GetDatabase(ctx context.Context, wallet, database string) (Database, error) {
	url, err := s.connector.GetConnectionInfo(wallet)
	if err != nil {
		return Database{}, err
	}

	client, err := sqlclient.Open(ctx, url)
	if err != nil {
		return Database{}, err
	}
	defer client.Close()

	targetOpts := &sqlspec.InspectRealmOption{Schemas: []string{database}}
	source, err := client.InspectRealm(ctx, targetOpts)
	if err != nil {
		return Database{}, err
	}

	schema, ok := source.Schema(database)
	if !ok {
		return Database{}, ErrDatabaseNotFound
	}

	return convertSchema(schema), nil
}
