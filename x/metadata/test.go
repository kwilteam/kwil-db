package metadata

import (
	"context"
	"ksl/postgres"
	"ksl/schema"
	"ksl/sqlschema"

	"github.com/google/uuid"
)

type dbinfo struct {
	Wallet   string
	Database string
}

type testservice struct {
	plans     map[uuid.UUID]PlanInfo
	databases map[dbinfo]sqlschema.Database
}

func NewTestService() Service {
	return &testservice{
		plans:     make(map[uuid.UUID]PlanInfo),
		databases: make(map[dbinfo]sqlschema.Database),
	}
}

func (s *testservice) Plan(ctx context.Context, req PlanRequest) (Plan, error) {
	ksch := schema.Parse(req.SchemaData, "<schema>")

	if ksch.HasErrors() {
		return Plan{}, ksch.Diagnostics
	}
	target := sqlschema.CalculateSqlSchema(ksch, req.Database)

	current, ok := s.databases[dbinfo{Wallet: req.Wallet, Database: req.Database}]
	if !ok {
		current = sqlschema.NewDatabase(req.Database)
	}

	differ := sqlschema.NewDiffer(postgres.Backend{})
	changes, err := differ.Diff(current, target)
	if err != nil {
		return Plan{}, err
	}

	planner := postgres.Planner{}
	plan, err := planner.Plan(sqlschema.Migration{Before: current, After: target, Changes: changes})
	if err != nil {
		return Plan{}, err
	}

	planID := uuid.New()
	s.plans[planID] = PlanInfo{
		Wallet:   req.Wallet,
		Database: req.Database,
		Data:     req.SchemaData,
	}

	return convertPlan(planID, plan), nil
}

func (s *testservice) Apply(ctx context.Context, planID uuid.UUID) error {
	plan, ok := s.plans[planID]
	if !ok {
		return ErrPlanNotFound
	}

	ksch := schema.Parse(plan.Data, "<schema>")

	if ksch.HasErrors() {
		return ksch.Diagnostics
	}

	target := sqlschema.CalculateSqlSchema(ksch, plan.Database)
	s.databases[dbinfo{Wallet: plan.Wallet, Database: plan.Database}] = target
	return nil
}

func (s *testservice) GetMetadata(ctx context.Context, req RequestMetadata) (Metadata, error) {
	db, ok := s.databases[dbinfo(req)]
	if !ok {
		return Metadata{}, ErrDatabaseNotFound
	}

	return convertDatabase(db), nil
}
