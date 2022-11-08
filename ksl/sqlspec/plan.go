package sqlspec

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
)

type (
	Plan struct {
		Version       string
		Name          string
		Reversible    bool
		Transactional bool
		Changes       []*Change
	}

	Change struct {
		Cmd     string
		Args    []any
		Comment string
		Reverse string
		Source  SchemaChange
	}
)

type (
	Planner interface {
		PlanChanges([]SchemaChange, ...PlanOption) (*Plan, error)
	}

	PlanApplier interface {
		ApplyChanges(context.Context, []SchemaChange, ...PlanOption) error
	}

	PlanOptions struct {
		SchemaQualifier *string
		Name            string
	}

	PlanOption func(*PlanOptions)
)

type ExecPlanner interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PlanChanges([]SchemaChange, ...PlanOption) (*Plan, error)
}

type ExecQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func ApplyChanges(ctx context.Context, changes []SchemaChange, p ExecPlanner, opts ...PlanOption) error {
	plan, err := p.PlanChanges(changes, opts...)
	if err != nil {
		return err
	}
	for _, c := range plan.Changes {
		if _, err := p.ExecContext(ctx, c.Cmd, c.Args...); err != nil {
			if c.Comment != "" {
				err = fmt.Errorf("%s: %w", c.Comment, err)
			}
			return err
		}
	}
	return nil
}

func DetachCycles(changes []SchemaChange) ([]SchemaChange, error) {
	sorted, err := sortMap(changes)
	if err == errCycle {
		return detachReferences(changes), nil
	}
	if err != nil {
		return nil, err
	}
	planned := make([]SchemaChange, len(changes))
	copy(planned, changes)
	sort.Slice(planned, func(i, j int) bool {
		return sorted[table(planned[i])] < sorted[table(planned[j])]
	})
	return planned, nil
}

func detachReferences(changes []SchemaChange) []SchemaChange {
	var planned, deferred []SchemaChange
	for _, change := range changes {
		switch change := change.(type) {
		case *AddTable:
			var (
				ext  []SchemaChange
				self []*ForeignKey
			)
			for _, fk := range change.T.ForeignKeys {
				if fk.RefTable == change.T {
					self = append(self, fk)
				} else {
					ext = append(ext, &AddForeignKey{F: fk})
				}
			}
			if len(ext) > 0 {
				deferred = append(deferred, &ModifyTable{T: change.T, Changes: ext})
				t := *change.T
				t.ForeignKeys = self
				change = &AddTable{T: &t, Extra: change.Extra}
			}
			planned = append(planned, change)
		case *DropTable:
			var fks []SchemaChange
			for _, fk := range change.T.ForeignKeys {
				if fk.RefTable != change.T {
					fks = append(fks, &DropForeignKey{F: fk})
				}
			}
			if len(fks) > 0 {
				planned = append(planned, &ModifyTable{T: change.T, Changes: fks})
				t := *change.T
				t.ForeignKeys = nil
				change = &DropTable{T: &t, Extra: change.Extra}
			}
			deferred = append(deferred, change)
		case *ModifyTable:
			var fks, rest []SchemaChange
			for _, c := range change.Changes {
				switch c := c.(type) {
				case *AddForeignKey:
					fks = append(fks, c)
				default:
					rest = append(rest, c)
				}
			}
			if len(fks) > 0 {
				deferred = append(deferred, &ModifyTable{T: change.T, Changes: fks})
			}
			if len(rest) > 0 {
				planned = append(planned, &ModifyTable{T: change.T, Changes: rest})
			}
		default:
			planned = append(planned, change)
		}
	}
	return append(planned, deferred...)
}

var errCycle = errors.New("cycle detected")

func sortMap(changes []SchemaChange) (map[string]int, error) {
	var (
		visit     func(string) bool
		sorted    = make(map[string]int)
		progress  = make(map[string]bool)
		deps, err = dependencies(changes)
	)
	if err != nil {
		return nil, err
	}
	visit = func(name string) bool {
		if _, done := sorted[name]; done {
			return false
		}
		if progress[name] {
			return true
		}
		progress[name] = true
		for _, ref := range deps[name] {
			if visit(ref.Name) {
				return true
			}
		}
		delete(progress, name)
		sorted[name] = len(sorted)
		return false
	}
	for node := range deps {
		if visit(node) {
			return nil, errCycle
		}
	}
	return sorted, nil
}

func dependencies(changes []SchemaChange) (map[string][]*Table, error) {
	deps := make(map[string][]*Table)
	for _, change := range changes {
		switch change := change.(type) {
		case *AddTable:
			for _, fk := range change.T.ForeignKeys {
				if err := checkFK(fk); err != nil {
					return nil, err
				}
				if fk.RefTable != change.T {
					deps[change.T.Name] = append(deps[change.T.Name], fk.RefTable)
				}
			}
		case *DropTable:
			for _, fk := range change.T.ForeignKeys {
				if err := checkFK(fk); err != nil {
					return nil, err
				}
				if isDropped(changes, fk.RefTable) {
					deps[fk.RefTable.Name] = append(deps[fk.RefTable.Name], fk.Table)
				}
			}
		case *ModifyTable:
			for _, c := range change.Changes {
				switch c := c.(type) {
				case *AddForeignKey:
					if err := checkFK(c.F); err != nil {
						return nil, err
					}
					if c.F.RefTable != change.T {
						deps[change.T.Name] = append(deps[change.T.Name], c.F.RefTable)
					}
				case *ModifyForeignKey:
					if err := checkFK(c.To); err != nil {
						return nil, err
					}
					if c.To.RefTable != change.T {
						deps[change.T.Name] = append(deps[change.T.Name], c.To.RefTable)
					}
				}
			}
		}
	}
	return deps, nil
}

func checkFK(fk *ForeignKey) error {
	var cause []string
	if fk.Table == nil {
		cause = append(cause, "child table")
	}
	if len(fk.Columns) == 0 {
		cause = append(cause, "child columns")
	}
	if fk.RefTable == nil {
		cause = append(cause, "parent table")
	}
	if len(fk.RefColumns) == 0 {
		cause = append(cause, "parent columns")
	}
	if len(cause) != 0 {
		return fmt.Errorf("missing %q for foreign key: %q", cause, fk.Name)
	}
	return nil
}

func table(change SchemaChange) (t string) {
	switch change := change.(type) {
	case *AddTable:
		t = change.T.Name
	case *DropTable:
		t = change.T.Name
	case *ModifyTable:
		t = change.T.Name
	}
	return
}

func isDropped(changes []SchemaChange, t *Table) bool {
	for _, c := range changes {
		if c, ok := c.(*DropTable); ok && c.T.Name == t.Name {
			return true
		}
	}
	return false
}

func CheckChangesScope(changes []SchemaChange) error {
	names := make(map[string]struct{})
	for _, c := range changes {
		var t *Table
		switch c := c.(type) {
		case *AddSchema, *ModifySchema, *DropSchema:
			return fmt.Errorf("%T is not allowed when migration plan is scoped to one schema", c)
		case *AddTable:
			t = c.T
		case *ModifyTable:
			t = c.T
		case *DropTable:
			t = c.T
		default:
			continue
		}
		if t.Schema != nil && t.Schema.Name != "" {
			names[t.Schema.Name] = struct{}{}
		}
		for _, c := range t.Columns {
			e, ok := c.Type.Type.(*EnumType)
			if ok && e.Schema != nil && e.Schema.Name != "" {
				names[t.Schema.Name] = struct{}{}
			}
		}
	}
	if len(names) > 1 {
		return fmt.Errorf("found %d schemas when migration plan is scoped to one", len(names))
	}
	return nil
}

func PrintPlan(p *Plan) {
	for _, c := range p.Changes {
		fmt.Println(c.Cmd)
	}
}
