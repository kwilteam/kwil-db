package sqlspec

import (
	"context"
)

type InspectMode uint

const (
	InspectSchemas InspectMode = 1 << iota
	InspectTables
)

func (m InspectMode) Is(i InspectMode) bool { return m&i != 0 }

type (
	InspectOptions struct {
		Mode    InspectMode
		Tables  []string
		Exclude []string
	}

	InspectRealmOption struct {
		Mode    InspectMode
		Schemas []string
		Exclude []string
	}
	Inspector interface {
		InspectSchema(ctx context.Context, name string, opts *InspectOptions) (*Schema, error)
		InspectRealm(ctx context.Context, opts *InspectRealmOption) (*Realm, error)
	}
)
