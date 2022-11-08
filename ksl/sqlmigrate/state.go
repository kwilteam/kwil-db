package sqlspec

import (
	"context"

	"ksl/sqlspec"
)

// StateReader wraps the method for reading a database/schema state.
type StateReader interface {
	ReadState(ctx context.Context) (*sqlspec.Realm, error)
}

// The StateReaderFunc type is an adapter to allow the use of
// ordinary functions as state readers.
type StateReaderFunc func(ctx context.Context) (*sqlspec.Realm, error)

func (f StateReaderFunc) ReadState(ctx context.Context) (*sqlspec.Realm, error) {
	return f(ctx)
}

// Realm returns a StateReader for the static Realm object.
func RealmReader(r *sqlspec.Realm) StateReader {
	return StateReaderFunc(func(context.Context) (*sqlspec.Realm, error) {
		return r, nil
	})
}

// RealmConn returns a StateReader for a Driver connected to a database.
func RealmConnReader(drv sqlspec.Driver, opts *sqlspec.InspectRealmOption) StateReader {
	return StateReaderFunc(func(ctx context.Context) (*sqlspec.Realm, error) {
		return drv.InspectRealm(ctx, opts)
	})
}

// Schema returns a StateReader for the static Schema object.
func SchemaReader(s *sqlspec.Schema) StateReader {
	return StateReaderFunc(func(context.Context) (*sqlspec.Realm, error) {
		r := &sqlspec.Realm{Schemas: []*sqlspec.Schema{s}}
		if s.Realm != nil {
			r.Attrs = s.Realm.Attrs
		}
		s.Realm = r
		return r, nil
	})
}

// SchemaConn returns a StateReader for a Driver connected to a schema.
func SchemaConnReader(drv sqlspec.Driver, name string, opts *sqlspec.InspectOptions) StateReader {
	return StateReaderFunc(func(ctx context.Context) (*sqlspec.Realm, error) {
		s, err := drv.InspectSchema(ctx, name, opts)
		if err != nil {
			return nil, err
		}
		return SchemaReader(s).ReadState(ctx)
	})
}
