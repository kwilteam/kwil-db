package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/internal/sqlx"
)

type (
	// A Plan defines a planned changeset that its execution brings the database to
	// the new desired state. Additional information is calculated by the different
	// drivers to indicate if the changeset is transactional (can be rolled-back) and
	// reversible (a down file can be generated to it).
	Plan struct {
		// Version and Name of the plan. Provided by the user or auto-generated.
		Version, Name string

		// Reversible describes if the changeset is reversible.
		Reversible bool

		// Transactional describes if the changeset is transactional.
		Transactional bool

		// Changes defines the list of changeset in the plan.
		Changes []*Change
	}

	// A Change of migration.
	Change struct {
		// Cmd or statement to execute.
		Cmd string

		// Args for placeholder parameters in the statement above.
		Args []any

		// A Comment describes the change.
		Comment string

		// Reverse contains the "reversed statement" if
		// command is reversible.
		Reverse string

		// The Source that caused this change, or nil.
		Source SchemaChange
	}
)

type (
	// The Driver interface must be implemented by the different dialects to support database
	// migration authoring/planning and applying. ExecQuerier, Inspector and Differ, provide
	// basic schema primitives for inspecting database schemas, calculate the difference between
	// schema elements, and executing raw SQL statements. The PlanApplier interface wraps the
	// methods for generating migration plan for applying the actual changes on the database.
	Driver interface {
		Differ
		sqlx.ExecQuerier
		Inspector
		Planner
		PlanApplier
	}

	// Planner wraps the methods for planning and applying changes
	// on the database.
	Planner interface {
		// PlanChanges returns a migration plan for applying the given changeset.
		PlanChanges(context.Context, []SchemaChange, ...PlanOption) (*Plan, error)
	}
	// PlanApplier wraps the methods for applying changes on the database.
	PlanApplier interface {
		// ApplyChanges is responsible for applying the given changeset.
		// An error may return from ApplyChanges if the driver is unable
		// to execute a change.
		ApplyChanges(context.Context, []SchemaChange, ...PlanOption) error
	}

	// PlanOptions holds the migration plan options to be used by PlanApplier.
	PlanOptions struct {
		// PlanWithSchemaQualifier allows setting a custom schema to prefix
		// tables and other resources. An empty string indicates no qualifier.
		SchemaQualifier *string
		Name            string
	}

	// PlanOption allows configuring a drivers' plan using functional arguments.
	PlanOption func(*PlanOptions)

	// StateReader wraps the method for reading a database/schema state.
	// The types below provides a few builtin options for reading a state
	// from a migration directory, a static object (e.g. a parsed file).
	StateReader interface {
		ReadState(ctx context.Context) (*Realm, error)
	}

	// The StateReaderFunc type is an adapter to allow the use of
	// ordinary functions as state readers.
	StateReaderFunc func(ctx context.Context) (*Realm, error)

	// Snapshotter wraps the Snapshot method.
	Snapshotter interface {
		// Snapshot takes a snapshot of the current database state and returns a function that can be called to restore
		// that state. Snapshot should return an error, if the current state can not be restored completely, e.g. if
		// there is a table already containing some rows.
		Snapshot(context.Context) (RestoreFunc, error)
	}

	// RestoreFunc is returned by the Snapshoter to explicitly restore the database state.
	RestoreFunc func(context.Context) error
)

// ReadState calls f(ctx).
func (f StateReaderFunc) ReadState(ctx context.Context) (*Realm, error) {
	return f(ctx)
}

// Realm returns a StateReader for the static Realm object.
func NewRealmReader(r *Realm) StateReader {
	return StateReaderFunc(func(context.Context) (*Realm, error) {
		return r, nil
	})
}

// Schema returns a StateReader for the static Schema object.
func NewSchemaReader(s *Schema) StateReader {
	return StateReaderFunc(func(context.Context) (*Realm, error) {
		r := &Realm{Schemas: []*Schema{s}}
		if s.Realm != nil {
			r.Attrs = s.Realm.Attrs
		}
		s.Realm = r
		return r, nil
	})
}

// RealmConn returns a StateReader for a Driver connected to a database.
func NewRealmConn(drv Driver, opts *InspectRealmOption) StateReader {
	return StateReaderFunc(func(ctx context.Context) (*Realm, error) {
		return drv.InspectRealm(ctx, opts)
	})
}

// SchemaConn returns a StateReader for a Driver connected to a
func NewSchemaConn(drv Driver, name string, opts *InspectOptions) StateReader {
	return StateReaderFunc(func(ctx context.Context) (*Realm, error) {
		s, err := drv.InspectSchema(ctx, name, opts)
		if err != nil {
			return nil, err
		}
		return NewSchemaReader(s).ReadState(ctx)
	})
}

type ExecPlanner interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PlanChanges(context.Context, []SchemaChange, ...PlanOption) (*Plan, error)
}

// ApplyChanges is a helper used by the different drivers to apply changes.
func ApplyChanges(ctx context.Context, changes []SchemaChange, p ExecPlanner, opts ...PlanOption) error {
	plan, err := p.PlanChanges(ctx, changes, opts...)
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

// DetachCycles takes a list of schema changes, and detaches
// references between changes if there is at least one circular
// reference in the changeset. More explicitly, it postpones fks
// creation, or deletes fks before deletes their tables.
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

// detachReferences detaches all table references.
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

// errCycle is an internal error to indicate a case of a cycle.
var errCycle = errors.New("cycle detected")

// sortMap returns an index-map indicates the position of table in a topological
// sort in reversed order based on its references, and a boolean indicate if there
// is a non-self loop.
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

// dependencies returned an adjacency list of all tables and the table they depend on
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
		return fmt.Errorf("missing %q for foreign key: %q", cause, fk.Symbol)
	}
	return nil
}

// table extracts a table from the given change.
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

// isDropped checks if the given table is marked as a deleted in the changeset.
func isDropped(changes []SchemaChange, t *Table) bool {
	for _, c := range changes {
		if c, ok := c.(*DropTable); ok && c.T.Name == t.Name {
			return true
		}
	}
	return false
}

// CheckChangesScope checks that changes can be applied
// on a schema scope (connection).
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
