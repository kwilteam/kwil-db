package schema

import (
	"context"
)

// An InspectMode controls the amount and depth of information returned on inspection.
type InspectMode uint

const (
	// InspectSchemas enables schema inspection.
	InspectSchemas InspectMode = 1 << iota

	// InspectTables enables schema tables inspection including
	// all its child resources (e.g. columns or indexes).
	InspectTables
)

// Is reports whether the given mode is enabled.
func (m InspectMode) Is(i InspectMode) bool { return m&i != 0 }

type (
	// InspectOptions describes options for Inspector.
	InspectOptions struct {
		// Mode defines the amount of information returned by InspectSchema.
		// If zero, InspectSchema inspects whole resources in the schema.
		Mode InspectMode

		// Tables to inspect. Empty means all tables in the schema.
		Tables []string

		// Exclude defines a list of glob patterns used to filter resources from inspection.
		// The syntax used by the different drivers is implemented as follows:
		//
		//	t   // exclude table 't'.
		//	*   // exclude all tables.
		//	t.c // exclude column, index and foreign-key named 'c' in table 't'.
		//	t.* // the last item defines the filtering; all resources under 't' are excluded.
		//	*.c // the last item defines the filtering; all resourced named 'c' are excluded in all tables.
		//	*.* // the last item defines the filtering; all resourced under all tables are excluded.
		//
		Exclude []string
	}

	// InspectRealmOption describes options for RealmInspector.
	InspectRealmOption struct {
		// Mode defines the amount of information returned by InspectRealm.
		// If zero, InspectRealm inspects all schemas and their child resources.
		Mode InspectMode

		// Schemas to inspect. Empty means all schemas in the realm.
		Schemas []string

		// Exclude defines a list of glob patterns used to filter resources from inspection.
		// The syntax used by the different drivers is implemented as follows:
		//
		//	s     // exclude schema 's'.
		//	*     // exclude all schemas.
		//	s.t   // exclude table 't' under schema 's'.
		//	s.*   // the last item defines the filtering; all tables under 's' are excluded.
		//	*.t   // the last item defines the filtering; all tables named 't' are excluded in all schemas.
		//	*.*   // the last item defines the filtering; all tables under all schemas are excluded.
		//	*.*.c // the last item defines the filtering; all resourced named 'c' are excluded in all tables.
		//	*.*.* // the last item defines the filtering; all resources are excluded in all tables.
		//
		Exclude []string
	}

	// Inspector is the interface implemented by the different database drivers for inspecting schema or databases.
	Inspector interface {
		// InspectSchema returns the schema description by its name. An empty name means the
		// "attached schema" (e.g. SCHEMA() in MySQL or CURRENT_SCHEMA() in PostgreSQL).
		// A NotExistError error is returned if the schema does not exist in the database.
		InspectSchema(ctx context.Context, name string, opts *InspectOptions) (*Schema, error)
		InspectRealm(ctx context.Context, opts *InspectRealmOption) (*Database, error)
	}
)
