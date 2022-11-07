package schema

import (
	"github.com/google/uuid"
	"github.com/kwilteam/ksl/sqlspec"
)

func convertSchema(r *sqlspec.Schema) Database {
	tables := make([]Table, len(r.Tables))
	for i, t := range r.Tables {
		tables[i] = convertTable(t)
	}
	enums := make([]Enum, len(r.Enums))
	for i, e := range r.Enums {
		enums[i] = convertEnum(e)
	}

	queries := make([]Query, len(r.Realm.Queries))
	for i, q := range r.Realm.Queries {
		queries[i] = Query{
			Name:      q.Name,
			Statement: q.Statement,
		}
	}

	roles := make([]Role, len(r.Realm.Roles))
	for i, r := range r.Realm.Roles {
		roles[i] = convertRole(r)
	}

	return Database{
		Name:    r.Name,
		Tables:  tables,
		Enums:   enums,
		Queries: queries,
		Roles:   roles,
	}
}

func convertRole(r *sqlspec.Role) Role {
	queries := make([]string, len(r.Queries))
	for i, q := range r.Queries {
		queries[i] = q.Name
	}

	return Role{
		Name:    r.Name,
		Queries: queries,
	}
}

func convertEnum(e *sqlspec.Enum) Enum {
	return Enum{
		Name:   e.Name,
		Values: e.Values[:],
	}
}

func convertTable(t *sqlspec.Table) Table {
	columns := make([]Column, len(t.Columns))
	for i, c := range t.Columns {
		columns[i] = convertColumn(c)
	}
	return Table{
		Name:    t.Name,
		Columns: columns,
	}
}

func convertColumn(c *sqlspec.Column) Column {
	typ := c.Type.Raw
	if t, err := sqlspec.TypeRegistry.Convert(c.Type.Type); err == nil {
		typ = t.Type
	}
	return Column{
		Name:     c.Name,
		Type:     typ,
		Nullable: c.Type.Nullable,
	}
}

func convertPlan(id uuid.UUID, p *sqlspec.Plan) Plan {
	changes := make([]Change, len(p.Changes))
	for i, c := range p.Changes {
		changes[i] = Change{
			Cmd:     c.Cmd,
			Args:    c.Args,
			Comment: c.Comment,
			Reverse: c.Reverse,
		}
	}

	return Plan{
		ID:            id,
		Version:       p.Version,
		Changes:       changes,
		Name:          p.Name,
		Reversible:    p.Reversible,
		Transactional: p.Transactional,
	}
}
