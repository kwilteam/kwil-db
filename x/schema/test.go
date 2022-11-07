package schema

import (
	"context"

	"github.com/google/uuid"
	"github.com/kwilteam/ksl/kslparse"
	"github.com/kwilteam/ksl/sqlspec"
)

type testservice struct {
	plans  map[uuid.UUID]PlanInfo
	realms map[string]*sqlspec.Realm
}

func NewTestService() Service {
	return &testservice{
		plans:  make(map[uuid.UUID]PlanInfo),
		realms: make(map[string]*sqlspec.Realm),
	}
}

func (s *testservice) Plan(ctx context.Context, req PlanRequest) (Plan, error) {
	parser := kslparse.NewParser()
	_, diags := parser.Parse(req.SchemaData, "<schema>")

	if diags.HasErrors() {
		return Plan{}, diags
	}

	target, diags := sqlspec.Decode(parser.FileSet())
	if diags.HasErrors() {
		return Plan{}, diags
	}

	current, ok := s.realms[req.Wallet]
	if !ok {
		current = &sqlspec.Realm{}
	}

	differ := sqlspec.NewDiffer()
	changes, err := differ.RealmDiff(current, target)
	if err != nil {
		return Plan{}, err
	}

	planner := sqlspec.NewPlanner()
	plan, err := planner.PlanChanges(changes)
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

	parser := kslparse.NewParser()
	_, diags := parser.Parse(plan.Data, "<schema>")

	if diags.HasErrors() {
		return diags
	}

	target, diags := sqlspec.Decode(parser.FileSet())
	if diags.HasErrors() {
		return diags
	}

	s.realms[plan.Wallet] = target
	return nil
}

func (s *testservice) GetDatabase(ctx context.Context, wallet, database string) (Database, error) {
	realm, ok := s.realms[wallet]
	if !ok {
		return Database{}, ErrDatabaseNotFound
	}

	schema, ok := realm.Schema(database)
	if !ok {
		return Database{}, ErrDatabaseNotFound
	}

	return convertSchema(schema), nil
}
