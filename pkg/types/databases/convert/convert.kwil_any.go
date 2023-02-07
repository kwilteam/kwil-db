package convert

import (
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

type kwilAnyConversion struct{}

// anytype.KwilAny ---> []byte

// DatabaseToBytes converts a Database[anytype.KwilAny] to Database[[]byte]
func (k *kwilAnyConversion) DatabaseToBytes(d *databases2.Database[anytype.KwilAny]) (*databases2.Database[[]byte], error) {
	// convert tables
	tables := make([]*databases2.Table[[]byte], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := k.TableToBytes(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	queries := make([]*databases2.SQLQuery[[]byte], len(d.SQLQueries))
	for i, q := range d.SQLQueries {
		qry := k.SQLQueryToBytes(q)
		queries[i] = qry
	}

	return &databases2.Database[[]byte]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: queries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToBytes converts a Table[anytype.KwilAny] to Table[[]byte]
func (k *kwilAnyConversion) TableToBytes(t *databases2.Table[anytype.KwilAny]) (*databases2.Table[[]byte], error) {
	// convert columns
	columns := make([]*databases2.Column[[]byte], len(t.Columns))
	for i, c := range t.Columns {
		col, err := k.ColumnToBytes(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &databases2.Table[[]byte]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToBytes converts a Column[anytype.KwilAny] to Column[[]byte]
func (k *kwilAnyConversion) ColumnToBytes(c *databases2.Column[anytype.KwilAny]) (*databases2.Column[[]byte], error) {
	// convert attributes
	attributes := make([]*databases2.Attribute[[]byte], len(c.Attributes))
	for i, a := range c.Attributes {
		value := a.Value.Bytes()

		attributes[i] = &databases2.Attribute[[]byte]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &databases2.Column[[]byte]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// SQLQueryToBytes converts a SQLQuery[anytype.KwilAny] to SQLQuery[[]byte]
func (k *kwilAnyConversion) SQLQueryToBytes(q *databases2.SQLQuery[anytype.KwilAny]) *databases2.SQLQuery[[]byte] {
	// convert parameters
	parameters := make([]*databases2.Parameter[[]byte], len(q.Params))
	for i, p := range q.Params {
		param := k.ParameterToBytes(p)
		parameters[i] = param
	}

	// convert where clauses
	whereClauses := make([]*databases2.WhereClause[[]byte], len(q.Where))
	for i, w := range q.Where {
		where := k.WhereClauseToBytes(w)
		whereClauses[i] = where
	}

	return &databases2.SQLQuery[[]byte]{
		Name:   q.Name,
		Type:   q.Type,
		Table:  q.Table,
		Params: parameters,
		Where:  whereClauses,
	}
}

// ParameterToBytes converts a Parameter[anytype.KwilAny] to Parameter[[]byte]
func (k *kwilAnyConversion) ParameterToBytes(p *databases2.Parameter[anytype.KwilAny]) *databases2.Parameter[[]byte] {
	return &databases2.Parameter[[]byte]{
		Name:     p.Name,
		Column:   p.Column,
		Static:   p.Static,
		Value:    p.Value.Bytes(),
		Modifier: p.Modifier,
	}
}

func (k *kwilAnyConversion) WhereClauseToBytes(w *databases2.WhereClause[anytype.KwilAny]) *databases2.WhereClause[[]byte] {
	return &databases2.WhereClause[[]byte]{
		Name:     w.Name,
		Column:   w.Column,
		Static:   w.Static,
		Operator: w.Operator,
		Value:    w.Value.Bytes(),
		Modifier: w.Modifier,
	}
}
