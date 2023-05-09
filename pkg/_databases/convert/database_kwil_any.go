package convert

import (
	"github.com/kwilteam/kwil-db/pkg/databases"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"
)

type kwilAnyConversion struct{}

// anytype.KwilAny ---> []byte

// DatabaseToBytes converts a Database[anytype.KwilAny] to Database[[]byte]
func (k *kwilAnyConversion) DatabaseToBytes(d *databases.Database[*spec.KwilAny]) (*databases.Database[[]byte], error) {
	// convert tables
	tables := make([]*databases.Table[[]byte], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := k.TableToBytes(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	queries := make([]*databases.SQLQuery[[]byte], len(d.SQLQueries))
	for i, q := range d.SQLQueries {
		qry := k.SQLQueryToBytes(q)
		queries[i] = qry
	}

	return &databases.Database[[]byte]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: queries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToBytes converts a Table[anytype.KwilAny] to Table[[]byte]
func (k *kwilAnyConversion) TableToBytes(t *databases.Table[*spec.KwilAny]) (*databases.Table[[]byte], error) {
	// convert columns
	columns := make([]*databases.Column[[]byte], len(t.Columns))
	for i, c := range t.Columns {
		col, err := k.ColumnToBytes(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &databases.Table[[]byte]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToBytes converts a Column[anytype.KwilAny] to Column[[]byte]
func (k *kwilAnyConversion) ColumnToBytes(c *databases.Column[*spec.KwilAny]) (*databases.Column[[]byte], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[[]byte], len(c.Attributes))
	for i, a := range c.Attributes {
		value := a.Value.Bytes()

		attributes[i] = &databases.Attribute[[]byte]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &databases.Column[[]byte]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// SQLQueryToBytes converts a SQLQuery[anytype.KwilAny] to SQLQuery[[]byte]
func (k *kwilAnyConversion) SQLQueryToBytes(q *databases.SQLQuery[*spec.KwilAny]) *databases.SQLQuery[[]byte] {
	// convert parameters
	parameters := make([]*databases.Parameter[[]byte], len(q.Params))
	for i, p := range q.Params {
		param := k.ParameterToBytes(p)
		parameters[i] = param
	}

	// convert where clauses
	whereClauses := make([]*databases.WhereClause[[]byte], len(q.Where))
	for i, w := range q.Where {
		where := k.WhereClauseToBytes(w)
		whereClauses[i] = where
	}

	return &databases.SQLQuery[[]byte]{
		Name:   q.Name,
		Type:   q.Type,
		Table:  q.Table,
		Params: parameters,
		Where:  whereClauses,
	}
}

// ParameterToBytes converts a Parameter[anytype.KwilAny] to Parameter[[]byte]
func (k *kwilAnyConversion) ParameterToBytes(p *databases.Parameter[*spec.KwilAny]) *databases.Parameter[[]byte] {
	return &databases.Parameter[[]byte]{
		Name:     p.Name,
		Column:   p.Column,
		Static:   p.Static,
		Value:    p.Value.Bytes(),
		Modifier: p.Modifier,
	}
}

func (k *kwilAnyConversion) WhereClauseToBytes(w *databases.WhereClause[*spec.KwilAny]) *databases.WhereClause[[]byte] {
	return &databases.WhereClause[[]byte]{
		Name:     w.Name,
		Column:   w.Column,
		Static:   w.Static,
		Operator: w.Operator,
		Value:    w.Value.Bytes(),
		Modifier: w.Modifier,
	}
}
