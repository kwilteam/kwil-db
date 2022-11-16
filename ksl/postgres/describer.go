package postgres

import (
	"context"
	"database/sql"
	"ksl"
	"ksl/sqldriver"
	"ksl/sqlschema"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

type Describer struct {
	Conn sqldriver.ExecQuerier
}

func (d Describer) Describe(schema string) (sqlschema.Database, error) {
	return d.DescribeContext(context.Background(), schema)
}

func (d Describer) DescribeContext(ctx context.Context, schema string) (sqlschema.Database, error) {
	sch := sqlschema.NewDatabase(schema)
	err := newSchemaCtx(d.Conn, &sch).describe(ctx)
	return sch, err
}

type schemactx struct {
	conn sqldriver.ExecQuerier
	db   *sqlschema.Database
}

func newSchemaCtx(conn sqldriver.ExecQuerier, db *sqlschema.Database) *schemactx {
	return &schemactx{conn: conn, db: db}
}

func (d *schemactx) describe(ctx context.Context) error {
	if err := d.loadTables(ctx); err != nil {
		return err
	}
	if err := d.loadEnums(ctx); err != nil {
		return err
	}
	if err := d.loadColumns(ctx); err != nil {
		return err
	}
	if err := d.loadForeignKeys(ctx); err != nil {
		return err
	}
	if err := d.loadIndexes(ctx); err != nil {
		return err
	}
	// if err := d.loadChecks(ctx); err != nil {
	// 	return err
	// }
	// if err := d.loadExtensions(ctx); err != nil {
	// 	return err
	// }

	return nil
}

func (d *schemactx) loadTables(ctx context.Context) error {
	query := `
	SELECT
		t1.table_name,
		pg_catalog.obj_description(t3.oid, 'pg_class') AS comment
	FROM
		INFORMATION_SCHEMA.TABLES AS t1
		JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
		JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
		LEFT JOIN pg_catalog.pg_partitioned_table AS t4 ON t4.partrelid = t3.oid
	WHERE
		t1.table_type = 'BASE TABLE'
		AND NOT COALESCE(t3.relispartition, false)
		AND t1.table_schema = $1
	ORDER BY
		t1.table_schema, t1.table_name;`

	rows, err := d.conn.QueryContext(ctx, query, d.db.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, comment sql.NullString
		if err := rows.Scan(&name, &comment); err != nil {
			return err
		}
		d.db.AddTable(sqlschema.Table{
			Name:    name.String,
			Comment: comment.String,
		})
	}

	return nil
}

func (d *schemactx) loadEnums(ctx context.Context) error {
	query := `
	SELECT t.typname as name, e.enumlabel as value
	FROM pg_type t
	JOIN pg_enum e ON t.oid = e.enumtypid
	JOIN pg_catalog.pg_namespace n ON n.oid = t.typnamespace
	WHERE n.nspname = $1
	ORDER BY e.enumsortorder;`

	rows, err := d.conn.QueryContext(ctx, query, d.db.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	enums := make(map[string][]string)

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			return err
		}
		enums[name] = append(enums[name], value)
	}

	for name, values := range enums {
		d.db.AddEnum(name, values...)
	}

	return nil
}

func (d *schemactx) loadColumns(ctx context.Context) error {
	query := `
	SELECT
		t1.table_name,
		t1.column_name,
		t1.data_type,
		pg_catalog.format_type(a.atttypid, a.atttypmod) AS format_type,
		t1.udt_name,
		(CASE WHEN t1.is_nullable = 'YES' THEN true ELSE false END) AS is_nullable,
		t1.column_default,
		t1.character_maximum_length,
		t1.numeric_precision,
		t1.datetime_precision,
		t1.numeric_scale,
		t1.interval_type,
		t1.character_set_name,
		t1.collation_name,
		(CASE WHEN t1.is_identity = 'YES' THEN true ELSE false END) AS is_identity,
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
		information_schema.columns AS t1
		JOIN pg_catalog.pg_namespace AS t2 ON t2.nspname = t1.table_schema
		JOIN pg_catalog.pg_class AS t3 ON t3.relnamespace = t2.oid AND t3.relname = t1.table_name
		JOIN pg_catalog.pg_attribute AS a ON a.attrelid = t3.oid AND a.attname = t1.column_name
		LEFT JOIN pg_catalog.pg_type AS t4 ON t1.udt_name = t4.typname AND t4.typnamespace = t2.oid
	WHERE
		t1.table_schema = $1
	ORDER BY
		t1.table_name, t1.ordinal_position;`

	rows, err := d.conn.QueryContext(ctx, query, d.db.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var col columndesc
		if err := rows.Scan(
			&col.TableName,
			&col.ColumnName,
			&col.DataType,
			&col.FormattedType,
			&col.TypeName,
			&col.IsNullable,
			&col.ColumnDefault,
			&col.CharacterMaximumLength,
			&col.NumericPrecision,
			&col.DateTimePrecision,
			&col.NumericScale,
			&col.IntervalType,
			&col.CharacterSetName,
			&col.CollationName,
			&col.IsIdentity,
			&col.IdentityStart,
			&col.IdentityIncrement,
			&col.IdentityLast,
			&col.IdentityGeneration,
			&col.GenerationExpression,
			&col.Comment,
			&col.TypeType,
			&col.TypeElem,
			&col.ElemType,
			&col.OID,
		); err != nil {
			return err
		}

		table, ok := d.db.FindTable(col.TableName).Get()
		if !ok {
			continue
		}

		var arity sqlschema.ColumnArity
		switch {
		case col.isArray():
			arity = sqlschema.List
		case col.IsNullable:
			arity = sqlschema.Nullable
		default:
			arity = sqlschema.Required
		}

		var nativeType ksl.Type
		typeName, typeArgs := col.nativeTypeArguments()

		if col.TypeType.String == "e" || col.ElemType.String == "e" || col.isUserDefined() {
			if enum, ok := d.db.FindEnum(typeName).Get(); ok {
				nativeType = sqlschema.EnumType{Name: enum.Name(), ID: enum.ID}
			} else {
				nativeType = ksl.UserDefinedType{Name: typeName}
			}
		} else {
			typ, err := ParseNativeType(typeName, itos(typeArgs)...)
			if err != nil {
				return err
			}
			nativeType = typ
		}

		d.db.AddColumn(sqlschema.Column{
			Table: table.ID,
			Name:  col.ColumnName,
			Type: sqlschema.ColumnType{
				Type:  nativeType,
				Arity: arity,
				Raw:   typeName,
			},
			Default:       nil,
			Comment:       col.Comment.String,
			Charset:       col.CharacterSetName.String,
			Collation:     col.CollationName.String,
			AutoIncrement: col.IsIdentity,
		})

		// if col.IsIdentity {
		// 	column.AddAttrs(&Identity{
		// 		Generation: col.IdentityGeneration.String,
		// 		Sequence: &Sequence{
		// 			Start:     int(col.IdentityStart.Int64),
		// 			Increment: int(col.IdentityIncrement.Int64),
		// 			Last:      int(col.IdentityLast.Int64),
		// 		},
		// 	})
		// }

		// if notNullOrEmpty(col.GenerationExpression) {
		// 	column.SetGeneratedExpression(col.GenerationExpression.String)
		// }
	}

	return nil
}

func (d *schemactx) loadForeignKeys(ctx context.Context) error {
	query := `
	SELECT
		t1.constraint_name,
		t1.table_name,
		t2.column_name,
		t3.table_name AS referenced_table_name,
		t3.column_name AS referenced_column_name,
		t4.update_rule,
		t4.delete_rule
	FROM
		information_schema.table_constraints t1
		JOIN information_schema.key_column_usage t2 ON t1.constraint_name = t2.constraint_name AND t1.table_schema = t2.constraint_schema
		JOIN information_schema.constraint_column_usage t3 ON t1.constraint_name = t3.constraint_name AND t1.table_schema = t3.constraint_schema
		JOIN information_schema.referential_constraints t4 ON t1.constraint_name = t4.constraint_name AND t1.table_schema = t4.constraint_schema
	WHERE
		t1.constraint_type = 'FOREIGN KEY'
		AND t1.table_schema = $1
		AND t1.table_name = ANY ( $2 )
	ORDER BY
		t1.constraint_name,
		t2.ordinal_position;`

	rows, err := d.conn.QueryContext(ctx, query, d.db.Name, d.db.TableNames())
	if err != nil {
		return err
	}
	defer rows.Close()

	currentFK := mo.None[lo.Tuple2[string, sqlschema.ForeignKeyID]]()

	for rows.Next() {
		var desc fkdesc
		if err := rows.Scan(
			&desc.Name,
			&desc.Table,
			&desc.Column,
			&desc.RefTable,
			&desc.RefColumn,
			&desc.UpdateRule,
			&desc.DeleteRule,
		); err != nil {
			return err
		}

		table := d.db.FindTable(desc.Table).MustGet()
		refTable := d.db.FindTable(desc.RefTable).MustGet()
		column := table.Column(desc.Column).MustGet()
		refColumn := refTable.Column(desc.RefColumn).MustGet()
		onDelete := foreignKeyAction(desc.DeleteRule)
		onUpdate := foreignKeyAction(desc.UpdateRule)

		if currentFK.IsAbsent() || currentFK.MustGet().A != desc.Name {
			fkid := d.db.AddForeignKey(sqlschema.ForeignKey{
				ConstraintName:   desc.Name,
				ConstrainedTable: table.ID,
				ReferencedTable:  refTable.ID,
				OnDeleteAction:   onDelete,
				OnUpdateAction:   onUpdate,
			})
			currentFK = mo.Some(lo.T2(desc.Name, fkid))
		}

		if currentFK.IsPresent() {
			d.db.AddForeignKeyColumn(sqlschema.ForeignKeyColumn{
				ForeignKey:        currentFK.MustGet().B,
				ConstrainedColumn: column.ID,
				ReferencedColumn:  refColumn.ID,
			})
		}
	}

	return nil
}

func (d *schemactx) loadIndexes(ctx context.Context) error {
	query := `
	WITH rawindex AS (
		SELECT
			indrelid,
			indexrelid,
			indisunique,
			indisprimary,
			indnatts,
			indnkeyatts,
			unnest(indkey) AS indkeyid,
			generate_subscripts(indkey, 1) AS indkeyidx,
			unnest(indclass) AS indclass,
			pg_get_expr(indpred, indrelid) AS predicate,
			unnest(indoption) AS indoption
		FROM pg_index
		WHERE
			indpred IS NULL
			AND array_position(indkey::int2[], 0::int2) IS NULL
	)
	SELECT
		indexinfo.relname AS index_name,
		tableinfo.relname AS table_name,
		columninfo.attname AS column_name,
		rawindex.indisunique AS is_unique,
		rawindex.indisprimary AS is_primary_key,
		rawindex.indkeyidx AS column_index,
		opclass.opcname AS opclass,
		opclass.opcdefault AS opcdefault,
		c.contype AS constraint_type,
		rawindex.predicate AS predicate,
		pg_get_indexdef(rawindex.indexrelid, rawindex.indkeyidx, false) AS expression,
		indexinfo.reloptions AS options,
		indexaccess.amname AS index_algo,
		obj_description(indexinfo.oid, 'pg_class') AS comment,
		(columninfo.attname <> '' AND rawindex.indnatts > rawindex.indnkeyatts AND rawindex.indkeyidx > rawindex.indnkeyatts) AS included,
		CASE rawindex.indoption & 1
			WHEN 1 THEN 'DESC'
			ELSE 'ASC' END
			AS column_order
	FROM
		rawindex
		INNER JOIN pg_class AS tableinfo ON tableinfo.oid = rawindex.indrelid
		INNER JOIN pg_class AS indexinfo ON indexinfo.oid = rawindex.indexrelid
		INNER JOIN pg_namespace AS schemainfo ON schemainfo.oid = tableinfo.relnamespace
		INNER JOIN pg_attribute AS columninfo ON columninfo.attrelid = tableinfo.oid AND columninfo.attnum = rawindex.indkeyid
		INNER JOIN pg_am AS indexaccess ON indexaccess.oid = indexinfo.relam
		LEFT JOIN pg_constraint AS c ON rawindex.indexrelid = c.conindid
		LEFT JOIN pg_opclass AS opclass ON opclass.oid = rawindex.indclass
	WHERE schemainfo.nspname = $1
	ORDER BY schemainfo.nspname, tableinfo.relname, indexinfo.relname, rawindex.indkeyidx;`

	rows, err := d.conn.QueryContext(ctx, query, d.db.Name)
	if err != nil {
		return err
	}
	defer rows.Close()

	currentIndex := mo.None[sqlschema.IndexID]()

	for rows.Next() {
		var index indexdesc
		if err := rows.Scan(
			&index.IndexName,
			&index.TableName,
			&index.ColumnName,
			&index.IsUnique,
			&index.IsPrimaryKey,
			&index.ColumnIndex,
			&index.OpClass,
			&index.OpcDefault,
			&index.ConstraintType,
			&index.Predicate,
			&index.Expression,
			&index.Options,
			&index.IndexAlgo,
			&index.Comment,
			&index.Included,
			&index.ColumnOrder,
		); err != nil {
			return err
		}

		table := d.db.FindTable(index.TableName).MustGet()

		indexAlgo := sqlschema.BTreeAlgo
		switch strings.ToUpper(index.IndexAlgo) {
		case "BTREE":
			indexAlgo = sqlschema.BTreeAlgo
		case "HASH":
			indexAlgo = sqlschema.HashAlgo
		case "GIST":
			indexAlgo = sqlschema.GistAlgo
		case "GIN":
			indexAlgo = sqlschema.GinAlgo
		case "BRIN":
			indexAlgo = sqlschema.BrinAlgo
		case "SPGIST":
			indexAlgo = sqlschema.SpGistAlgo
		}

		if index.ColumnIndex == 0 {
			switch {
			case index.IsPrimaryKey:
				currentIndex = mo.Some(d.db.AddPrimaryKey(sqlschema.Index{Table: table.ID, Name: index.IndexName, Algorithm: indexAlgo}))
			case index.IsUnique:
				currentIndex = mo.Some(d.db.AddUniqueIndex(sqlschema.Index{Table: table.ID, Name: index.IndexName, Algorithm: indexAlgo}))
			default:
				currentIndex = mo.Some(d.db.AddIndex(sqlschema.Index{Table: table.ID, Name: index.IndexName, Algorithm: indexAlgo}))
			}
		}

		column := table.Column(index.ColumnName.String).MustGet()
		var sortOrder sqlschema.SortOrder
		switch strings.ToUpper(index.ColumnOrder) {
		case "ASC":
			sortOrder = sqlschema.Ascending
		case "DESC":
			sortOrder = sqlschema.Descending
		}

		d.db.AddIndexColumn(sqlschema.IndexColumn{
			Index:     currentIndex.MustGet(),
			Column:    column.ID,
			SortOrder: sortOrder,
		})
	}

	return nil
}

// func (d *schemactx) loadChecks(ctx context.Context) error {
// 	query := `
// 	SELECT
// 		rel.relname AS table_name,
// 		t1.conname AS constraint_name,
// 		pg_get_expr(t1.conbin, t1.conrelid) as expression,
// 		t2.attname as column_name,
// 		t1.conkey as column_indexes,
// 		t1.connoinherit as no_inherit
// 	FROM
// 		pg_constraint t1
// 		JOIN pg_attribute t2
// 		ON t2.attrelid = t1.conrelid
// 		AND t2.attnum = ANY (t1.conkey)
// 		JOIN pg_class rel
// 		ON rel.oid = t1.conrelid
// 		JOIN pg_namespace nsp
// 		ON nsp.oid = t1.connamespace
// 	WHERE
// 		t1.contype = 'c'
// 		AND nsp.nspname = $1
// 		AND rel.relname = ANY ( $2)
// 	ORDER BY
// 		t1.conname, array_position(t1.conkey, t2.attnum);`

// 	rows, err := d.conn.QueryContext(ctx, query, d.db.Name, d.db.TableNames())
// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var desc checkdesc
// 		if err := rows.Scan(&desc.TableName, &desc.ConstraintName, &desc.Clause, &desc.ColumnName, &desc.ColumnIndexes, &desc.NoInherit); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (d *schemactx) loadExtensions(ctx context.Context) error {
// 	query := `
// 	SELECT
// 		ext.extname AS extension_name,
// 		ext.extversion AS extension_version,
// 		ext.extrelocatable AS extension_relocatable,
// 		pn.nspname AS extension_schema
// 	FROM pg_extension ext
// 	INNER JOIN pg_namespace pn ON ext.extnamespace = pn.oid
// 	ORDER BY ext.extname ASC;`

// 	rows, err := d.conn.QueryContext(ctx, query)
// 	if err != nil {
// 		return err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var e sqlschema.Extension
// 		if err := rows.Scan(&e.Name, &e.Version, &e.Relocatable, &e.Namespace); err != nil {
// 			return err
// 		}
// 		d.db.AddExtension(e)
// 	}

// 	return nil
// }

var precSingle = regexp.MustCompile(`.*\(([0-9]*)\).*\[\]$`)
var precDual = regexp.MustCompile(`\w+\(([0-9]*),([0-9]*)\)\[\]$`)

//	type checkdesc struct {
//		TableName      string
//		ConstraintName string
//		ColumnName     string
//		Clause         string
//		ColumnIndexes  []int32
//		NoInherit      bool
//	}
type fkdesc struct {
	Name       string
	Table      string
	Column     string
	RefTable   string
	RefColumn  string
	UpdateRule string
	DeleteRule string
}

type indexdesc struct {
	IndexName      string
	TableName      string
	ColumnName     sql.NullString
	IsUnique       bool
	IsPrimaryKey   bool
	ColumnIndex    int
	OpClass        string
	OpcDefault     bool
	IndexAlgo      string
	ConstraintType sql.NullString
	Predicate      sql.NullString
	Expression     sql.NullString
	Options        sql.NullString
	Comment        sql.NullString
	Included       bool
	ColumnOrder    string
}

type columndesc struct {
	TableName              string
	ColumnName             string
	DataType               string
	FormattedType          string
	TypeName               string
	IsNullable             bool
	ColumnDefault          sql.NullString
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	DateTimePrecision      sql.NullInt64
	NumericScale           sql.NullInt64
	IntervalType           sql.NullString
	CharacterSetName       sql.NullString
	CollationName          sql.NullString
	IsIdentity             bool
	IdentityStart          sql.NullInt64
	IdentityIncrement      sql.NullInt64
	IdentityLast           sql.NullInt64
	IdentityGeneration     sql.NullString
	GenerationExpression   sql.NullString
	Comment                sql.NullString
	TypeType               sql.NullString
	TypeElem               sql.NullInt64
	ElemType               sql.NullString
	OID                    sql.NullInt64
}

func (c columndesc) isArray() bool { return strings.EqualFold(c.DataType, TypeArray) }
func (c columndesc) isUserDefined() bool {
	return strings.EqualFold(c.DataType, TypeUserDefined)
}

func (c columndesc) nativeTypeArguments() (string, []int) {
	if c.isArray() {
		typeName := strings.TrimPrefix(c.TypeName, "_")
		if matches := precSingle.FindStringSubmatch(c.FormattedType); len(matches) > 1 {
			return typeName, stoi(matches[1:2])
		} else if matches := precDual.FindStringSubmatch(c.FormattedType); len(matches) > 2 {
			return typeName, stoi(matches[1:3])
		}
		return typeName, nil
	}

	typeName := c.TypeName
	var args []int
	switch AliasType(typeName) {
	case TypeNumeric:
		switch {
		case c.NumericPrecision.Valid && c.NumericScale.Valid:
			args = append(args, int(c.NumericPrecision.Int64), int(c.NumericScale.Int64))
		case c.NumericPrecision.Valid:
			args = append(args, int(c.NumericPrecision.Int64))
		}
	case TypeTimestamp, TypeTimestampTZ, TypeTime, TypeTimeTZ, TypeInterval:
		if c.DateTimePrecision.Valid {
			args = append(args, int(c.DateTimePrecision.Int64))
		}
	case TypeBit, TypeVarBit, TypeChar, TypeVarChar:
		if c.CharacterMaximumLength.Valid {
			args = append(args, int(c.CharacterMaximumLength.Int64))
		}
	}
	return typeName, args
}

const (
	NoActionString   string = "NO ACTION"
	RestrictString   string = "RESTRICT"
	CascadeString    string = "CASCADE"
	SetNullString    string = "SET NULL"
	SetDefaultString string = "SET DEFAULT"
)

func foreignKeyAction(action string) sqlschema.ForeignKeyAction {
	switch action {
	case NoActionString:
		return sqlschema.NoAction
	case RestrictString:
		return sqlschema.Restrict
	case CascadeString:
		return sqlschema.Cascade
	case SetNullString:
		return sqlschema.SetNull
	case SetDefaultString:
		return sqlschema.SetDefault
	default:
		return sqlschema.NoAction
	}
}
