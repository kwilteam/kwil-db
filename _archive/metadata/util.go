package metadata

import (
	"ksl"
	"ksl/postgres"
	"ksl/sqlmigrate"
	"ksl/sqlschema"
)

type metadatadb struct {
	db   sqlschema.Database
	meta *Metadata
}

func (m *metadatadb) convert() Metadata {
	m.meta.DbName = m.db.Name
	m.convertTables()
	m.convertEnums()
	m.generateQueries()
	return *m.meta
}

func (m *metadatadb) convertTables() {
	tables := make([]Table, len(m.db.Tables))
	for i, table := range m.db.WalkTables() {
		tables[i] = Table{
			Name: table.Name(),
		}
	}
	m.meta.Tables = tables
}

func (m *metadatadb) convertEnums() {
	enums := make([]Enum, len(m.db.Enums))
	for i, e := range m.db.WalkEnums() {
		enums[i] = Enum{Name: e.Name()}
	}
	m.meta.Enums = enums
}

func (m *metadatadb) generateQueries() {
	for _, table := range m.db.WalkTables() {
		m.generateGetQuery(table)
		m.generateCreateQuery(table)
		m.generateUpdateQuery(table)
		m.generateDeleteQuery(table)
	}
}

func (m *metadatadb) generateGetQuery(table sqlschema.TableWalker) {
	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return
	}

	q := Query{Name: "Get" + table.Name()}
	for _, field := range pk.Columns() {
		q.Inputs = append(q.Inputs, convertColumnToParam(field.Column()))
	}

	for _, column := range table.Columns() {
		q.Outputs = append(q.Outputs, convertColumnToParam(column))
	}

	m.meta.Queries = append(m.meta.Queries, q)
}

func (m *metadatadb) generateCreateQuery(table sqlschema.TableWalker) {
	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return
	}

	q := Query{Name: "Create" + table.Name()}

	for _, column := range table.Columns() {
		q.Inputs = append(q.Inputs, convertColumnToParam(column))
	}

	for _, field := range pk.Columns() {
		q.Outputs = append(q.Outputs, convertColumnToParam(field.Column()))
	}

	m.meta.Queries = append(m.meta.Queries, q)
}

func (m *metadatadb) generateUpdateQuery(table sqlschema.TableWalker) {
	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return
	}

	q := Query{Name: "Update" + table.Name()}
	for _, field := range pk.Columns() {
		q.Inputs = append(q.Inputs, Param{
			Name:  field.Column().Name(),
			Arity: Required,
			Type:  convertSqlType(field.Column().Type()),
		})
	}

	for _, column := range table.Columns() {
		arity := Optional
		if pk.ContainsColumn(column.ID) {
			arity = Required
		}
		q.Inputs = append(q.Inputs, Param{
			Name:  column.Name(),
			Arity: arity,
			Type:  convertSqlType(column.Type()),
		})
	}

	m.meta.Queries = append(m.meta.Queries, q)
}

func (m *metadatadb) generateDeleteQuery(table sqlschema.TableWalker) {
	pk, ok := table.PrimaryKey().Get()
	if !ok {
		return
	}
	q := Query{Name: "Delete" + table.Name()}
	for _, field := range pk.Columns() {
		q.Inputs = append(q.Inputs, Param{
			Name:  field.Column().Name(),
			Arity: Required,
			Type:  convertSqlType(field.Column().Type()),
		})
	}

	m.meta.Queries = append(m.meta.Queries, q)
}

func convertDatabase(db sqlschema.Database) Metadata {
	return (&metadatadb{db: db, meta: &Metadata{}}).convert()
}

func convertPlan(p sqlmigrate.MigrationPlan) Plan {
	changes := make([]Change, len(p.Statements))
	for i, c := range p.Statements {
		changes[i] = Change{
			Cmd:     c.String(),
			Comment: c.Comment,
		}
	}

	return Plan{
		Changes: changes,
	}
}

func convertColumnToParam(c sqlschema.ColumnWalker) Param {
	return Param{
		Name:  c.Name(),
		Arity: convertSqlArity(c.Arity()),
		Type:  convertSqlType(c.Type()),
	}
}
func convertSqlArity(arity sqlschema.ColumnArity) TypeArity {
	switch arity {
	case sqlschema.Nullable:
		return Optional
	case sqlschema.Required:
		return Required
	case sqlschema.List:
		return Repeated
	default:
		return Optional
	}
}

func convertSqlType(t sqlschema.ColumnType) string {
	typ := postgres.ScalarTypeForNativeType(t.Type)
	switch typ {
	case ksl.BuiltIns.String:
		return ScalarString
	case ksl.BuiltIns.BigInt, ksl.BuiltIns.Float, ksl.BuiltIns.Int, ksl.BuiltIns.Decimal:
		return ScalarNumber
	case ksl.BuiltIns.Bool:
		return ScalarBool
	case ksl.BuiltIns.Date:
		return ScalarDate
	case ksl.BuiltIns.Time:
		return ScalarTime
	case ksl.BuiltIns.DateTime:
		return ScalarDateTime
	case ksl.BuiltIns.Bytes:
		return ScalarBytes
	default:
		return ScalarString
	}
}
