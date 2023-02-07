package convert

import (
	"fmt"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

type bytesConversion struct{}

// []byte ---> anytype.KwilAny
func (b *bytesConversion) DatabaseToKwilAny(d *databases2.Database[[]byte]) (*databases2.Database[anytype.KwilAny], error) {
	// convert tables
	tables := make([]*databases2.Table[anytype.KwilAny], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := b.TableToKwilAny(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	queries := make([]*databases2.SQLQuery[anytype.KwilAny], len(d.SQLQueries))
	for i, q := range d.SQLQueries {
		qry, err := b.SQLQueryToKwilAny(q)
		if err != nil {
			return nil, err
		}
		queries[i] = qry
	}

	return &databases2.Database[anytype.KwilAny]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: queries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToKwilAny converts a Table[[]byte] to Table[anytype.KwilAny]
func (b *bytesConversion) TableToKwilAny(t *databases2.Table[[]byte]) (*databases2.Table[anytype.KwilAny], error) {
	// convert columns
	columns := make([]*databases2.Column[anytype.KwilAny], len(t.Columns))
	for i, c := range t.Columns {
		col, err := b.ColumnToKwilAny(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &databases2.Table[anytype.KwilAny]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToKwilAny converts a Column[[]byte] to Column[anytype.KwilAny]
func (b *bytesConversion) ColumnToKwilAny(c *databases2.Column[[]byte]) (*databases2.Column[anytype.KwilAny], error) {
	// convert attributes
	attributes := make([]*databases2.Attribute[anytype.KwilAny], len(c.Attributes))
	for i, a := range c.Attributes {
		attr, err := anytype.NewFromSerial(a.Value)
		if err != nil {
			return nil, err
		}
		attributes[i] = &databases2.Attribute[anytype.KwilAny]{
			Type:  a.Type,
			Value: attr,
		}
	}

	return &databases2.Column[anytype.KwilAny]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// SQLQueryToKwilAny converts a SQLQuery[[]byte] to SQLQuery[anytype.KwilAny]
func (b *bytesConversion) SQLQueryToKwilAny(q *databases2.SQLQuery[[]byte]) (*databases2.SQLQuery[anytype.KwilAny], error) {
	// convert query
	params := make([]*databases2.Parameter[anytype.KwilAny], len(q.Params))
	for i, p := range q.Params {
		param, err := anytype.NewFromSerial(p.Value)
		if err != nil {
			return nil, fmt.Errorf("error converting parameter %s: %w", p.Name, err)
		}
		params[i] = &databases2.Parameter[anytype.KwilAny]{
			Name:     p.Name,
			Column:   p.Column,
			Static:   p.Static,
			Value:    param,
			Modifier: p.Modifier,
		}
	}

	wheres := make([]*databases2.WhereClause[anytype.KwilAny], len(q.Where))
	for i, w := range q.Where {
		where, err := anytype.NewFromSerial(w.Value)
		if err != nil {
			panic(err)
		}
		wheres[i] = &databases2.WhereClause[anytype.KwilAny]{
			Name:     w.Name,
			Column:   w.Column,
			Static:   w.Static,
			Operator: w.Operator,
			Value:    where,
			Modifier: w.Modifier,
		}
	}

	return &databases2.SQLQuery[anytype.KwilAny]{
		Name:   q.Name,
		Type:   q.Type,
		Table:  q.Table,
		Params: params,
		Where:  wheres,
	}, nil
}
