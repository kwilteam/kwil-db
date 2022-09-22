package schema

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"
)

// DevDriver is a driver that provides additional functionality
// to interact with the development database.
type DevDriver struct {
	// A Driver connected to the dev database.
	Driver

	// MaxNameLen configures the max length of object names in
	// the connected database (e.g. 64 in MySQL). Longer names
	// are trimmed and suffixed with their hash.
	MaxNameLen int

	// DropClause holds optional clauses that
	// can be added to the DropSchema change.
	DropClause []SchemaClause

	// PatchColumn allows providing a custom function to patch
	// columns that hold a schema reference.
	PatchColumn func(*Schema, *Column)
}

// NormalizeRealm implements the Normalizer interface.
//
// The implementation converts schema objects in "natural form" (e.g. HCL or DSL)
// to their "normal presentation" in the database, by creating them temporarily in
// a "dev database", and then inspects them from there.
func (d *DevDriver) NormalizeRealm(ctx context.Context, r *Realm) (nr *Realm, err error) {
	var (
		names   = make(map[string]string)
		changes = make([]SchemaChange, 0, len(r.Schemas))
		reverse = make([]SchemaChange, 0, len(r.Schemas))
		opts    = &InspectRealmOption{
			Schemas: make([]string, 0, len(r.Schemas)),
		}
	)
	for _, s := range r.Schemas {
		if s.Realm != r {
			s.Realm = r
		}
		dev := d.formatName(s.Name)
		names[dev] = s.Name
		s.Name = dev
		opts.Schemas = append(opts.Schemas, s.Name)
		// Skip adding the IfNotExists clause
		// to fail if the schema exists.
		st := New(dev).AddAttrs(s.Attrs...)
		changes = append(changes, &AddSchema{S: st})
		reverse = append(reverse, &DropSchema{S: st, Extra: append(d.DropClause, &IfExists{})})
		for _, t := range s.Tables {
			// If objects are not strongly connected.
			if t.Schema != s {
				t.Schema = s
			}
			for _, c := range t.Columns {
				if e, ok := c.Type.Type.(*EnumType); ok && e.Schema != s {
					e.Schema = s
				}
				if d.PatchColumn != nil {
					d.PatchColumn(s, c)
				}
			}
			changes = append(changes, &AddTable{T: t})
		}
	}
	patch := func(r *Realm) {
		for _, s := range r.Schemas {
			s.Name = names[s.Name]
		}
	}
	// Delete the dev resources, and return
	// the source realm to its initial state.
	defer func() {
		patch(r)
		if rerr := d.ApplyChanges(ctx, reverse); rerr != nil {
			if err != nil {
				rerr = fmt.Errorf("%w: %v", err, rerr)
			}
			err = rerr
		}
	}()
	if err := d.ApplyChanges(ctx, changes); err != nil {
		return nil, err
	}
	if nr, err = d.InspectRealm(ctx, opts); err != nil {
		return nil, err
	}
	patch(nr)
	return nr, nil
}

// NormalizeSchema returns the normal representation of the given database. See NormalizeRealm for more info.
func (d *DevDriver) NormalizeSchema(ctx context.Context, s *Schema) (*Schema, error) {
	r := &Realm{}
	if s.Realm != nil {
		r.Attrs = s.Realm.Attrs
	}
	r.Schemas = append(r.Schemas, s)
	nr, err := d.NormalizeRealm(ctx, r)
	if err != nil {
		return nil, err
	}
	ns, ok := nr.Schema(s.Name)
	if !ok {
		return nil, fmt.Errorf("missing normalized schema %q", s.Name)
	}
	return ns, nil
}

func (d *DevDriver) formatName(name string) string {
	dev := fmt.Sprintf("atlas_dev_%s_%d", name, time.Now().Unix())
	if d.MaxNameLen == 0 || len(dev) <= d.MaxNameLen {
		return dev
	}
	h := fnv.New128()
	h.Write([]byte(dev))
	return fmt.Sprintf("%s_%x", dev[:d.MaxNameLen-1-h.Size()*2], h.Sum(nil))
}
