package sqlspec

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"ksl/sqlutil"
)

type inspect struct {
	ExecQuerier
	ctype   string
	collate string
	version int
}

var _ Inspector = (*inspect)(nil)

func NewInspector(conn ExecQuerier, version int, charset, collate string) Inspector {
	return &inspect{ExecQuerier: conn, ctype: charset, collate: collate, version: version}
}

func (i *inspect) InspectRealm(ctx context.Context, opts *InspectRealmOption) (*Realm, error) {
	schemas, err := i.schemas(ctx, opts)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &InspectRealmOption{}
	}
	r := NewRealm(schemas...).SetCollation(i.collate)
	r.Attrs = append(r.Attrs, &CType{Value: i.ctype})
	if len(schemas) == 0 || !ModeInspectRealm(opts).Is(InspectTables) {
		return ExcludeRealm(r, opts.Exclude)
	}
	if err := i.inspectRoleQueries(ctx, r); err != nil {
		return nil, err
	}

	if err := i.inspectTables(ctx, r, nil); err != nil {
		return nil, err
	}
	LinkSchemaTables(schemas)
	return ExcludeRealm(r, opts.Exclude)
}

func (i *inspect) InspectSchema(ctx context.Context, name string, opts *InspectOptions) (s *Schema, err error) {
	schemas, err := i.schemas(ctx, &InspectRealmOption{Schemas: []string{name}})
	if err != nil {
		return nil, err
	}
	switch n := len(schemas); {
	case n == 0:
		return nil, &sqlutil.NotExistError{Err: fmt.Errorf("postgres: schema %q was not found", name)}
	case n > 1:
		return nil, fmt.Errorf("postgres: %d schemas were found for %q", n, name)
	}
	if opts == nil {
		opts = &InspectOptions{}
	}
	r := NewRealm(schemas...).SetCollation(i.collate)
	r.Attrs = append(r.Attrs, &CType{Value: i.ctype})
	if ModeInspectSchema(opts).Is(InspectTables) {
		if err := i.inspectRoleQueries(ctx, r); err != nil {
			return nil, err
		}

		if err := i.inspectTables(ctx, r, opts); err != nil {
			return nil, err
		}
		LinkSchemaTables(schemas)
	}
	return ExcludeSchema(r.Schemas[0], opts.Exclude)
}

func (i *inspect) inspectRoles(ctx context.Context, r *Realm) error {
	rows, err := i.QueryContext(ctx, rolesQuery)
	if err != nil {
		return fmt.Errorf("postgres: querying roles: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var name string
		var isDefault bool

		if err := rows.Scan(&name, &isDefault); err != nil {
			return fmt.Errorf("postgres: scanning role: %w", err)
		}

		r.GetOrCreateRole(name).SetDefault(isDefault)
	}

	return rows.Close()
}

func (i *inspect) inspectQueries(ctx context.Context, r *Realm) error {
	rows, err := i.QueryContext(ctx, queriesQuery)
	if err != nil {
		return fmt.Errorf("postgres: querying roles: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var name, statement string

		if err := rows.Scan(&name, &statement); err != nil {
			return fmt.Errorf("postgres: scanning role: %w", err)
		}

		r.GetOrCreateQuery(name).SetStatement(statement)
	}

	return rows.Close()
}

func (i *inspect) inspectRoleQueries(ctx context.Context, r *Realm) error {
	if err := i.inspectQueries(ctx, r); err != nil {
		return err
	}
	if err := i.inspectRoles(ctx, r); err != nil {
		return err
	}

	rows, err := i.QueryContext(ctx, roleQueriesQuery)
	if err != nil {
		return fmt.Errorf("postgres: querying role queries: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var roleName string
		var queryNames string

		if err := rows.Scan(&roleName, &queryNames); err != nil {
			return fmt.Errorf("postgres: scanning role query: %w", err)
		}
		role := r.GetOrCreateRole(roleName)
		queries := make([]*Query, len(queryNames))
		for i, queryName := range strings.Split(queryNames, ",") {
			queries[i] = r.GetOrCreateQuery(queryName)
		}
		role.AddQueries(queries...)
	}

	return rows.Close()
}

func (i *inspect) inspectTables(ctx context.Context, r *Realm, opts *InspectOptions) error {
	if err := i.tables(ctx, r, opts); err != nil {
		return err
	}
	for _, s := range r.Schemas {
		if len(s.Tables) == 0 {
			continue
		}
		if err := i.enums(ctx, s); err != nil {
			return err
		}
		if err := i.columns(ctx, s); err != nil {
			return err
		}
		if err := i.indexes(ctx, s); err != nil {
			return err
		}
		if err := i.partitions(s); err != nil {
			return err
		}
		if err := i.fks(ctx, s); err != nil {
			return err
		}
		if err := i.checks(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspect) enums(ctx context.Context, s *Schema) error {
	enumsQuery := `
select n.nspname as enum_schema,
	t.typname as enum_name,
	e.enumlabel as enum_value
from pg_type t
join pg_enum e on t.oid = e.enumtypid
join pg_catalog.pg_namespace n ON n.oid = t.typnamespace
where n.nspname = $1
order by e.oid
`

	rows, err := i.QueryContext(ctx, enumsQuery, s.Name)
	if err != nil {
		return fmt.Errorf("postgres: querying enums: %w", err)
	}

	defer rows.Close()
	for rows.Next() {
		var schema, name, value string
		if err := rows.Scan(&schema, &name, &value); err != nil {
			return fmt.Errorf("postgres: scanning enum: %w", err)
		}
		enum := s.GetOrCreateEnum(name)
		enum.Values = append(enum.Values, value)
	}

	return rows.Close()
}

func (i *inspect) tables(ctx context.Context, realm *Realm, opts *InspectOptions) error {
	var (
		args  []any
		query = fmt.Sprintf(tablesQuery, nArgs(0, len(realm.Schemas)))
	)
	for _, s := range realm.Schemas {
		args = append(args, s.Name)
	}
	if opts != nil && len(opts.Tables) > 0 {
		for _, t := range opts.Tables {
			args = append(args, t)
		}
		query = fmt.Sprintf(tablesQueryArgs, nArgs(0, len(realm.Schemas)), nArgs(len(realm.Schemas), len(opts.Tables)))
	}
	rows, err := i.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tSchema, name, comment, partattrs, partstart, partexprs sql.NullString
		if err := rows.Scan(&tSchema, &name, &comment, &partattrs, &partstart, &partexprs); err != nil {
			return fmt.Errorf("scan table information: %w", err)
		}
		if !sqlutil.ValidString(tSchema) || !sqlutil.ValidString(name) {
			return fmt.Errorf("invalid schema or table name: %q.%q", tSchema.String, name.String)
		}
		s, ok := realm.Schema(tSchema.String)
		if !ok {
			return fmt.Errorf("schema %q was not found in realm", tSchema.String)
		}
		t := &Table{Name: name.String}
		s.AddTables(t)
		if sqlutil.ValidString(comment) {
			t.SetComment(comment.String)
		}
		if sqlutil.ValidString(partattrs) {
			t.AddAttrs(&Partition{
				start: partstart.String,
				attrs: partattrs.String,
				exprs: partexprs.String,
			})
		}
	}
	return rows.Close()
}

func (i *inspect) columns(ctx context.Context, s *Schema) error {
	query := columnsQuery
	rows, err := i.querySchema(ctx, query, s)
	if err != nil {
		return fmt.Errorf("postgres: querying schema %q columns: %w", s.Name, err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := i.addColumn(s, rows); err != nil {
			return fmt.Errorf("postgres: %w", err)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if err := i.enumValues(ctx, s); err != nil {
		return err
	}
	return nil
}

func (i *inspect) addColumn(s *Schema, rows *sql.Rows) (err error) {
	var (
		typid, typelem, maxlen, precision, timeprecision, scale, seqstart, seqinc, seqlast                                                  sql.NullInt64
		table, name, typ, fmtype, nullable, defaults, identity, genidentity, genexpr, charset, collate, comment, typtype, elemtyp, interval sql.NullString
	)
	if err = rows.Scan(
		&table, &name, &typ, &fmtype, &nullable, &defaults, &maxlen, &precision, &timeprecision, &scale, &interval, &charset,
		&collate, &identity, &seqstart, &seqinc, &seqlast, &genidentity, &genexpr, &comment, &typtype, &typelem, &elemtyp, &typid,
	); err != nil {
		return err
	}
	t, ok := s.Table(table.String)
	if !ok {
		return fmt.Errorf("table %q was not found in schema", table.String)
	}
	colType, err := columnType(&columnDesc{
		typ:           typ.String,
		fmtype:        fmtype.String,
		size:          maxlen.Int64,
		scale:         scale.Int64,
		typtype:       typtype.String,
		typelem:       typelem.Int64,
		elemtyp:       elemtyp.String,
		typid:         typid.Int64,
		interval:      interval.String,
		precision:     precision.Int64,
		timePrecision: &timeprecision.Int64,
	})
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}

	c := NewColumn(name.String).SetColumnType(colType, typ.String, nullable.String == "YES")

	if defaults.Valid {
		defaultExpr(c, defaults.String)
	}
	if identity.String == "YES" {
		c.AddAttrs(&Identity{
			Generation: genidentity.String,
			Sequence: &Sequence{
				Last:      seqlast.Int64,
				Start:     seqstart.Int64,
				Increment: seqinc.Int64,
			},
		})
	}
	if sqlutil.ValidString(genexpr) {
		c.AddAttrs(&GeneratedExpr{Expr: genexpr.String})
	}
	if sqlutil.ValidString(comment) {
		c.SetComment(comment.String)
	}
	if sqlutil.ValidString(charset) {
		c.SetCharset(charset.String)
	}
	if sqlutil.ValidString(collate) {
		c.SetCollation(collate.String)
	}
	t.AddColumns(c)
	return nil
}

func (i *inspect) enumValues(ctx context.Context, s *Schema) error {
	var (
		args  []any
		ids   = make(map[int64][]*EnumType)
		query = "SELECT enumtypid, enumlabel FROM pg_enum WHERE enumtypid IN (%s)"
		newE  = func(e1 *enumType) *EnumType {
			if _, ok := ids[e1.ID]; !ok {
				args = append(args, e1.ID)
			}
			e2 := &EnumType{T: e1.T, Schema: s}
			if e1.Schema != "" && e1.Schema != s.Name {
				e2.Schema = New(e1.Schema)
			}
			ids[e1.ID] = append(ids[e1.ID], e2)
			return e2
		}
	)
	for _, t := range s.Tables {
		for _, c := range t.Columns {
			switch t := c.Type.Type.(type) {
			case *enumType:
				e := newE(t)
				c.Type.Type = e
				c.Type.Raw = e.T
			case *ArrayType:
				if e, ok := t.Type.(*enumType); ok {
					t.Type = newE(e)
				}
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	rows, err := i.QueryContext(ctx, fmt.Sprintf(query, nArgs(0, len(args))), args...)
	if err != nil {
		return fmt.Errorf("postgres: querying enum values: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id int64
			v  string
		)
		if err := rows.Scan(&id, &v); err != nil {
			return fmt.Errorf("postgres: scanning enum label: %w", err)
		}
		for _, enum := range ids[id] {
			enum.Values = append(enum.Values, v)
		}
	}
	return nil
}

func (i *inspect) indexes(ctx context.Context, s *Schema) error {
	query := indexesQuery
	if !i.supportsIndexInclude() {
		query = indexesQueryNoInclude
	}

	rows, err := i.querySchema(ctx, query, s)
	if err != nil {
		return fmt.Errorf("postgres: querying schema %q indexes: %w", s.Name, err)
	}
	defer rows.Close()
	if err := i.addIndexes(s, rows); err != nil {
		return err
	}
	return rows.Err()
}

func (i *inspect) supportsIndexInclude() bool {
	return i.version >= 11_00_00
}

func (i *inspect) addIndexes(s *Schema, rows *sql.Rows) error {
	names := make(map[string]*Index)
	for rows.Next() {
		var (
			uniq, primary, included                       bool
			table, name, typ                              string
			desc, nullsfirst, nullslast                   sql.NullBool
			column, contype, pred, expr, comment, options sql.NullString
		)
		if err := rows.Scan(&table, &name, &typ, &column, &included, &primary, &uniq, &contype, &pred, &expr, &desc, &nullsfirst, &nullslast, &comment, &options); err != nil {
			return fmt.Errorf("postgres: scanning indexes for schema %q: %w", s.Name, err)
		}
		t, ok := s.Table(table)
		if !ok {
			return fmt.Errorf("table %q was not found in schema", table)
		}
		idx, ok := names[name]
		if !ok {
			idx = NewIndex(name).SetUnique(uniq)
			var attrs []Attr = []Attr{&IndexType{T: strings.ToUpper(typ)}}

			if sqlutil.ValidString(comment) {
				attrs = append(attrs, &Comment{Text: comment.String})
			}
			if sqlutil.ValidString(contype) {
				attrs = append(attrs, &ConstraintType{Type: contype.String})
			}
			if sqlutil.ValidString(pred) {
				attrs = append(attrs, &IndexPredicate{Predicate: pred.String})
			}
			if sqlutil.ValidString(options) {
				p, err := newIndexStorage(options.String)
				if err != nil {
					return err
				}
				attrs = append(attrs, p)
			}
			idx.AddAttrs(attrs...)
			if primary {
				t.SetPrimaryKey(idx)
			} else {
				t.AddIndexes(idx)
			}

			names[name] = idx
		}
		part := NewIndexPart().SetDesc(desc.Bool)
		if nullsfirst.Bool || nullslast.Bool {
			part.AddAttrs(&IndexColumnProperty{
				NullsFirst: nullsfirst.Bool,
				NullsLast:  nullslast.Bool,
			})
		}
		switch {
		case included:
			c, ok := t.Column(column.String)
			if !ok {
				return fmt.Errorf("postgres: INCLUDE column %q was not found for index %q", column.String, idx.Name)
			}
			var include IndexInclude
			has(idx.Attrs, &include)
			include.Columns = append(include.Columns, c.Name)
			replaceOrAppend(&idx.Attrs, &include)
		case sqlutil.ValidString(column):
			col, ok := t.Column(column.String)
			if !ok {
				return fmt.Errorf("postgres: column %q was not found for index %q", column.String, idx.Name)
			}
			idx.AddParts(part.SetColumn(col))
		case sqlutil.ValidString(expr):
			idx.AddParts(part.SetExpr(&RawExpr{Expr: expr.String}))
		default:
			return fmt.Errorf("postgres: invalid part for index %q", idx.Name)
		}
	}
	return nil
}

func (i *inspect) partitions(s *Schema) error {
	for _, t := range s.Tables {
		var d Partition
		if !has(t.Attrs, &d) {
			continue
		}
		switch s := strings.ToLower(d.start); s {
		case "r":
			d.T = PartitionTypeRange
		case "l":
			d.T = PartitionTypeList
		case "h":
			d.T = PartitionTypeHash
		default:
			return fmt.Errorf("postgres: unexpected partition strategy %q", s)
		}
		idxs := strings.Split(strings.TrimSpace(d.attrs), " ")
		if len(idxs) == 0 {
			return fmt.Errorf("postgres: no columns/expressions were found in partition key for column %q", t.Name)
		}
		for i := range idxs {
			switch idx, err := strconv.Atoi(idxs[i]); {
			case err != nil:
				return fmt.Errorf("postgres: faild parsing partition key index %q", idxs[i])
			// An expression.
			case idx == 0:
				j := sqlutil.ExprLastIndex(d.exprs)
				if j == -1 {
					return fmt.Errorf("postgres: no expression found in partition key: %q", d.exprs)
				}
				d.Parts = append(d.Parts, &PartitionPart{
					Expr: &RawExpr{Expr: d.exprs[:j+1]},
				})
				d.exprs = strings.TrimPrefix(d.exprs[j+1:], ", ")
			// A column at index idx-1.
			default:
				if idx > len(t.Columns) {
					return fmt.Errorf("postgres: unexpected column index %d", idx)
				}
				d.Parts = append(d.Parts, &PartitionPart{
					Column: t.Columns[idx-1].Name,
				})
			}
		}
		replaceOrAppend(&t.Attrs, &d)
	}
	return nil
}

func (i *inspect) fks(ctx context.Context, s *Schema) error {
	rows, err := i.querySchema(ctx, fksQuery, s)
	if err != nil {
		return fmt.Errorf("postgres: querying schema %q foreign keys: %w", s.Name, err)
	}
	defer rows.Close()
	if err := SchemaFKs(s, rows); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	return rows.Err()
}

func (i *inspect) checks(ctx context.Context, s *Schema) error {
	rows, err := i.querySchema(ctx, checksQuery, s)
	if err != nil {
		return fmt.Errorf("postgres: querying schema %q check constraints: %w", s.Name, err)
	}
	defer rows.Close()
	if err := i.addChecks(s, rows); err != nil {
		return err
	}
	return rows.Err()
}

func (i *inspect) addChecks(s *Schema, rows *sql.Rows) error {
	names := make(map[string]*Check)
	for rows.Next() {
		var (
			noInherit                            bool
			table, name, column, clause, indexes string
		)
		if err := rows.Scan(&table, &name, &clause, &column, &indexes, &noInherit); err != nil {
			return fmt.Errorf("postgres: scanning check: %w", err)
		}
		t, ok := s.Table(table)
		if !ok {
			return fmt.Errorf("table %q was not found in schema", table)
		}
		if _, ok := t.Column(column); !ok {
			return fmt.Errorf("postgres: column %q was not found for check %q", column, name)
		}
		check, ok := names[name]
		if !ok {
			check = &Check{Name: name, Expr: clause, Attrs: []Attr{&CheckColumns{}}}
			if noInherit {
				check.Attrs = append(check.Attrs, &NoInherit{})
			}
			names[name] = check
			t.Attrs = append(t.Attrs, check)
		}
		c := check.Attrs[0].(*CheckColumns)
		c.Columns = append(c.Columns, column)
	}
	return nil
}

func (i *inspect) schemas(ctx context.Context, opts *InspectRealmOption) ([]*Schema, error) {
	var (
		args  []any
		query = schemasQuery
	)
	if opts != nil {
		switch n := len(opts.Schemas); {
		case n == 1 && opts.Schemas[0] == "":
			query = fmt.Sprintf(schemasQueryArgs, "= CURRENT_SCHEMA()")
		case n == 1 && opts.Schemas[0] != "":
			query = fmt.Sprintf(schemasQueryArgs, "= $1")
			args = append(args, opts.Schemas[0])
		case n > 0:
			query = fmt.Sprintf(schemasQueryArgs, "IN ("+nArgs(0, len(opts.Schemas))+")")
			for _, s := range opts.Schemas {
				args = append(args, s)
			}
		}
	}

	rows, err := i.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: querying schemas: %w", err)
	}
	defer rows.Close()

	var schemas []*Schema
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		schemas = append(schemas, &Schema{Name: name})
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return schemas, nil
}

func (i *inspect) querySchema(ctx context.Context, query string, s *Schema) (*sql.Rows, error) {
	args := []any{s.Name}
	for _, t := range s.Tables {
		args = append(args, t.Name)
	}
	return i.QueryContext(ctx, fmt.Sprintf(query, nArgs(1, len(s.Tables))), args...)
}

func nArgs(start, n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		if i > 1 {
			b.WriteString(", ")
		}
		b.WriteByte('$')
		b.WriteString(strconv.Itoa(start + i))
	}
	return b.String()
}

var reNextval = regexp.MustCompile(`(?i) *nextval\('(?:[\w$]+\.)*([\w$]+_[\w$]+_seq)'(?:::regclass)*\) *$`)

func defaultExpr(c *Column, s string) {
	switch m := reNextval.FindStringSubmatch(s); {
	case len(m) == 2:
		tt, ok := c.Type.Type.(*IntegerType)
		if !ok {
			return
		}
		st := &SerialType{SequenceName: m[1]}
		st.SetType(tt)
		c.Type.Raw = st.T
		c.Type.Type = st
	case sqlutil.IsLiteralBool(s), sqlutil.IsLiteralNumber(s), sqlutil.IsQuoted(s, '\''):
		c.Default = &LiteralExpr{Value: s}
	default:
		var x Expr = &RawExpr{Expr: s}
		// Try casting or fallback to raw expressions (e.g. column text[] has the default of '{}':text[]).
		if v, ok := canConvert(c.Type, s); ok {
			x = &LiteralExpr{Value: v}
		}
		c.Default = x
	}
}

func canConvert(t *ColumnType, x string) (string, bool) {
	i := strings.LastIndex(x, "::")
	if i == -1 || !sqlutil.IsQuoted(x[:i], '\'') {
		return "", false
	}
	q := x[0:i]
	x = x[1 : i-1]
	switch t.Type.(type) {
	case *EnumType:
		return q, true
	case *BoolType:
		if sqlutil.IsLiteralBool(x) {
			return x, true
		}
	case *DecimalType, *IntegerType, *FloatType:
		if sqlutil.IsLiteralNumber(x) {
			return x, true
		}
	case *ArrayType, *BinaryType, *JSONType, *NetworkType, *SpatialType, *StringType, *TimeType, *UUIDType, *XMLType:
		return q, true
	}
	return "", false
}

const (
	queriesQuery     = `SELECT name, statement FROM kwil.queries`
	rolesQuery       = `SELECT name, is_default FROM kwil.roles`
	roleQueriesQuery = `
SELECT
	r.name as role_name, string_agg(distinct q.name, ',') as query_names
FROM
	kwil.role_queries as rq
	JOIN kwil.roles as r ON r.id = rq.role_id
	JOIN kwil.queries as q ON q.id = rq.query_id
GROUP BY r.name;
`

	// Query to list database schemas.
	schemasQuery = "SELECT schema_name FROM information_schema.schemata WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast', 'crdb_internal', 'pg_extension', 'kwil') AND schema_name NOT LIKE 'pg_%temp_%' ORDER BY schema_name"

	// Query to list specific database schemas.
	schemasQueryArgs = "SELECT schema_name FROM information_schema.schemata WHERE schema_name %s ORDER BY schema_name"

	// Query to list table information.
	tablesQuery = `
SELECT
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
ORDER BY
	t1.table_schema, t1.table_name
`
	tablesQueryArgs = `
SELECT
	t1.table_schema,
	t1.table_name,
	pg_catalog.obj_description(t3.oid, 'pg_class') AS comment,
	t4.partattrs AS partition_attrs,
	t4.partstrat AS partition_strategy,
	pg_get_expr(t4.partexprs, t4.partrelid) AS partition_exprs
FROM
	INFORMATION_SCHEMA.TABLES AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
WHERE
	t1.table_type = 'BASE TABLE'
	AND NOT COALESCE(t3.relispartition, false)
	AND t1.table_schema IN (%s)
	AND t1.table_name IN (%s)
ORDER BY
	t1.table_schema, t1.table_name
`
	// Query to list table columns.
	columnsQuery = `
SELECT
	t1.table_name,
	t1.column_name,
	t1.data_type,
	pg_catalog.format_type(a.atttypid, a.atttypmod) AS format_type,
	t1.is_nullable,
	t1.column_default,
	t1.character_maximum_length,
	t1.numeric_precision,
	t1.datetime_precision,
	t1.numeric_scale,
	t1.interval_type,
	t1.character_set_name,
	t1.collation_name,
	t1.is_identity,
	t1.identity_start,
	t1.identity_increment,
	(CASE WHEN t1.is_identity = 'YES' THEN (SELECT last_value FROM pg_sequences WHERE quote_ident(schemaname) || '.' || quote_ident(sequencename) = pg_get_serial_sequence(quote_ident(t1.table_schema) || '.' || quote_ident(t1.table_name), t1.column_name)) END) AS identity_last,
	t1.identity_generation,
	t1.generation_expression,
	col_description(t3.oid, "ordinal_position") AS comment,
	t4.typtype,
	t4.typelem,
	(CASE WHEN t4.typcategory = 'A' AND t4.typelem <> 0 THEN (SELECT t.typtype FROM pg_catalog.pg_type t WHERE t.oid = t4.typelem) END) AS elemtyp,
	t4.oid
FROM
	"information_schema"."columns" AS t1
	JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
	JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
	JOIN pg_catalog.pg_attribute AS a ON a.attrelid = t3.oid AND a.attname = t1.column_name
	LEFT JOIN pg_catalog.pg_type AS t4 ON t1.udt_name = t4.typname AND t4.typnamespace = t2.oid
WHERE
	t1.table_schema = $1 AND t1.table_name IN (%s)
ORDER BY
	t1.table_name, t1.ordinal_position
`
	fksQuery = `
SELECT
    t1.constraint_name,
    t1.table_name,
    t2.column_name,
    t1.table_schema,
    t3.table_name AS referenced_table_name,
    t3.column_name AS referenced_column_name,
    t3.table_schema AS referenced_schema_name,
    t4.update_rule,
    t4.delete_rule
FROM
    information_schema.table_constraints t1
    JOIN information_schema.key_column_usage t2
    ON t1.constraint_name = t2.constraint_name
    AND t1.table_schema = t2.constraint_schema
    JOIN information_schema.constraint_column_usage t3
    ON t1.constraint_name = t3.constraint_name
    AND t1.table_schema = t3.constraint_schema
    JOIN information_schema.referential_constraints t4
    ON t1.constraint_name = t4.constraint_name
    AND t1.table_schema = t4.constraint_schema
WHERE
    t1.constraint_type = 'FOREIGN KEY'
    AND t1.table_schema = $1
    AND t1.table_name IN (%s)
ORDER BY
    t1.constraint_name,
    t2.ordinal_position
`

	// Query to list table check constraints.
	checksQuery = `
SELECT
	rel.relname AS table_name,
	t1.conname AS constraint_name,
	pg_get_expr(t1.conbin, t1.conrelid) as expression,
	t2.attname as column_name,
	t1.conkey as column_indexes,
	t1.connoinherit as no_inherit
FROM
	pg_constraint t1
	JOIN pg_attribute t2
	ON t2.attrelid = t1.conrelid
	AND t2.attnum = ANY (t1.conkey)
	JOIN pg_class rel
	ON rel.oid = t1.conrelid
	JOIN pg_namespace nsp
	ON nsp.oid = t1.connamespace
WHERE
	t1.contype = 'c'
	AND nsp.nspname = $1
	AND rel.relname IN (%s)
ORDER BY
	t1.conname, array_position(t1.conkey, t2.attnum)
`
)

var (
	indexesQuery          = fmt.Sprintf(indexesQueryTmpl, "(a.attname <> '' AND idx.indnatts > idx.indnkeyatts AND idx.ord > idx.indnkeyatts)", "%s")
	indexesQueryNoInclude = fmt.Sprintf(indexesQueryTmpl, "false", "%s")
	indexesQueryTmpl      = `
SELECT
	t.relname AS table_name,
	i.relname AS index_name,
	am.amname AS index_type,
	a.attname AS column_name,
	%s AS included,
	idx.indisprimary AS primary,
	idx.indisunique AS unique,
	c.contype AS constraint_type,
	pg_get_expr(idx.indpred, idx.indrelid) AS predicate,
	pg_get_indexdef(idx.indexrelid, idx.ord, false) AS expression,
	pg_index_column_has_property(idx.indexrelid, idx.ord, 'desc') AS desc,
	pg_index_column_has_property(idx.indexrelid, idx.ord, 'nulls_first') AS nulls_first,
	pg_index_column_has_property(idx.indexrelid, idx.ord, 'nulls_last') AS nulls_last,
	obj_description(i.oid, 'pg_class') AS comment,
	i.reloptions AS options
FROM
	(
		select
			*,
			generate_series(1,array_length(i.indkey,1)) as ord,
			unnest(i.indkey) AS key
		from pg_index i
	) idx
	JOIN pg_class i ON i.oid = idx.indexrelid
	JOIN pg_class t ON t.oid = idx.indrelid
	JOIN pg_namespace n ON n.oid = t.relnamespace
	LEFT JOIN pg_constraint c ON idx.indexrelid = c.conindid
	LEFT JOIN pg_attribute a ON (a.attrelid, a.attnum) = (idx.indrelid, idx.key)
	JOIN pg_am am ON am.oid = i.relam
WHERE
	n.nspname = $1
	AND t.relname IN (%s)
	AND COALESCE(c.contype, '') <> 'f'
ORDER BY
	table_name, index_name, idx.ord
`

	// functionsQuery = `
	// select n.nspname as schema_name,
	//        p.proname as specific_name,
	//        case p.prokind
	//             when 'f' then 'FUNCTION'
	//             when 'p' then 'PROCEDURE'
	//             when 'a' then 'AGGREGATE'
	//             when 'w' then 'WINDOW'
	//             end as kind,
	//        l.lanname as language,
	//        case when l.lanname = 'internal' then p.prosrc
	//             else pg_get_functiondef(p.oid)
	//             end as definition,
	//        pg_get_function_arguments(p.oid) as arguments,
	//        t.typname as return_type
	// from pg_proc p
	// left join pg_namespace n on p.pronamespace = n.oid
	// left join pg_language l on p.prolang = l.oid
	// left join pg_type t on t.oid = p.prorettype
	// where n.nspname not in ('pg_catalog', 'information_schema')
	// order by schema_name,
	//          specific_name;
	// `

	//	rolesQuery = `
	//
	// SELECT r.rolname, r.rolsuper, r.rolinherit, r.rolcreaterole, r.rolcreatedb, r.rolcanlogin, r.rolconnlimit, r.rolvaliduntil, r.rolreplication, r.rolbypassrls,
	//
	//	ARRAY(SELECT b.rolname
	//		FROM pg_catalog.pg_auth_members m
	//		JOIN pg_catalog.pg_roles b ON (m.roleid = b.oid)
	//		WHERE m.member = r.oid) as memberof
	//
	// FROM pg_catalog.pg_roles r
	// WHERE r.rolname !~ '^pg_'
	// ORDER BY 1;
	// `
)
