package sqlmigrate

import (
	"context"
	"ksl/sqlschema"
	"strings"
)

type Migration struct {
	Before, After sqlschema.Database
	Changes       []MigrationStep
}

type Migrator interface {
	PlanMigration(ctx context.Context, before, after sqlschema.Database) (MigrationPlan, error)
	ApplyMigration(context.Context, MigrationPlan) error
}

type MigrationPlan struct {
	Statements []Statement
}

func (m MigrationPlan) String() string {
	var b strings.Builder
	for _, stmt := range m.Statements {
		b.WriteString(stmt.String())
	}
	return b.String()
}

type Statement struct {
	Steps   Steps
	Comment string
}

func (s Statement) String() string {
	var b strings.Builder
	for _, step := range s.Steps {
		if step.Comment != "" {
			lines := strings.Split(step.Comment, "\n")
			for i := range lines {
				b.WriteString("-- " + lines[i] + "\n")
			}
		}
		b.WriteString(step.Cmd + ";\n")
	}
	return b.String()
}

type Steps []Step

func (s *Steps) Add(steps ...Step) { *s = append(*s, steps...) }

type Step struct {
	Cmd     string
	Args    []any
	Comment string
}

type Planner interface {
	Plan(Migration) (MigrationPlan, error)
	PlanContext(context.Context, Migration) (MigrationPlan, error)
}
