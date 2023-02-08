package convert

import (
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

type bytesConversion struct{}

// []byte ---> anytype.KwilAny
func (b *bytesConversion) DatabaseToKwilAny(d *databases.Database[[]byte]) (*databases.Database[*spec.KwilAny], error) {
	// convert tables
	tables := make([]*databases.Table[*spec.KwilAny], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := b.TableToKwilAny(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	queries := make([]*databases.SQLQuery[*spec.KwilAny], len(d.SQLQueries))
	for i, q := range d.SQLQueries {
		qry, err := b.SQLQueryToKwilAny(q)
		if err != nil {
			return nil, err
		}
		queries[i] = qry
	}

	return &databases.Database[*spec.KwilAny]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: queries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToKwilAny converts a Table[[]byte] to Table[anytype.KwilAny]
func (b *bytesConversion) TableToKwilAny(t *databases.Table[[]byte]) (*databases.Table[*spec.KwilAny], error) {
	// convert columns
	columns := make([]*databases.Column[*spec.KwilAny], len(t.Columns))
	for i, c := range t.Columns {
		col, err := b.ColumnToKwilAny(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &databases.Table[*spec.KwilAny]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToKwilAny converts a Column[[]byte] to Column[anytype.KwilAny]
func (b *bytesConversion) ColumnToKwilAny(c *databases.Column[[]byte]) (*databases.Column[*spec.KwilAny], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[*spec.KwilAny], len(c.Attributes))
	for i, a := range c.Attributes {
		attr, err := spec.NewFromSerial(a.Value)
		if err != nil {
			return nil, err
		}
		attributes[i] = &databases.Attribute[*spec.KwilAny]{
			Type:  a.Type,
			Value: attr,
		}
	}

	return &databases.Column[*spec.KwilAny]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// SQLQueryToKwilAny converts a SQLQuery[[]byte] to SQLQuery[anytype.KwilAny]
func (b *bytesConversion) SQLQueryToKwilAny(q *databases.SQLQuery[[]byte]) (*databases.SQLQuery[*spec.KwilAny], error) {
	// convert query
	params := make([]*databases.Parameter[*spec.KwilAny], len(q.Params))
	for i, p := range q.Params {
		param, err := spec.NewFromSerial(p.Value)
		if err != nil {
			return nil, fmt.Errorf("error converting parameter %s: %w", p.Name, err)
		}
		params[i] = &databases.Parameter[*spec.KwilAny]{
			Name:     p.Name,
			Column:   p.Column,
			Static:   p.Static,
			Value:    param,
			Modifier: p.Modifier,
		}
	}

	wheres := make([]*databases.WhereClause[*spec.KwilAny], len(q.Where))
	for i, w := range q.Where {
		where, err := spec.NewFromSerial(w.Value)
		if err != nil {
			panic(err)
		}
		wheres[i] = &databases.WhereClause[*spec.KwilAny]{
			Name:     w.Name,
			Column:   w.Column,
			Static:   w.Static,
			Operator: w.Operator,
			Value:    where,
			Modifier: w.Modifier,
		}
	}

	return &databases.SQLQuery[*spec.KwilAny]{
		Name:   q.Name,
		Type:   q.Type,
		Table:  q.Table,
		Params: params,
		Where:  wheres,
	}, nil
}
