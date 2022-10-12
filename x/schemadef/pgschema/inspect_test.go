package pgschema

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"unicode"

	"kwil/x/schemadef/sqlschema"
	"kwil/x/sql/sqlutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func rows(table string) *sqlmock.Rows {
	var (
		nc    int
		rows  *sqlmock.Rows
		lines = strings.Split(table, "\n")
	)
	for i := 0; i < len(lines); i++ {
		line := strings.TrimFunc(lines[i], unicode.IsSpace)
		// Skip new lines, header and footer.
		if line == "" || strings.IndexAny(line, "+-") == 0 {
			continue
		}
		columns := strings.FieldsFunc(line, func(r rune) bool {
			return r == '|'
		})
		for i, c := range columns {
			columns[i] = strings.TrimSpace(c)
		}
		if rows == nil {
			nc = len(columns)
			rows = sqlmock.NewRows(columns)
		} else {
			values := make([]driver.Value, nc)
			for i, c := range columns {
				switch c {
				case "", "nil", "NULL":
				default:
					values[i] = c
				}
			}
			rows.AddRow(values...)
		}
	}
	return rows
}

// Single table queries used by the different tests.
var (
	queryFKs     = sqlutil.Escape(fmt.Sprintf(fksQuery, "$2"))
	queryTables  = sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1"))
	queryChecks  = sqlutil.Escape(fmt.Sprintf(checksQuery, "$2"))
	queryColumns = sqlutil.Escape(fmt.Sprintf(columnsQuery, "$2"))
	queryIndexes = sqlutil.Escape(fmt.Sprintf(indexesQuery, "$2"))
)

func TestDriver_InspectTable(t *testing.T) {
	tests := []struct {
		name   string
		before func(mock)
		expect func(*require.Assertions, *sqlschema.Table, error)
	}{
		{
			name: "column types",
			before: func(m mock) {
				m.tableExists("public", "users", true)
				m.ExpectQuery(queryColumns).
					WithArgs("public", "users").
					WillReturnRows(rows(`
 table_name  |  column_name |          data_type          |  formatted          | is_nullable |         column_default                 | character_maximum_length | numeric_precision | datetime_precision | numeric_scale |    interval_type    | character_set_name | collation_name | is_identity | identity_start | identity_increment |   identity_last  | identity_generation | generation_expression | comment | typtype | typelem | elemtyp |  oid
-------------+--------------+-----------------------------+---------------------|-------------+----------------------------------------+--------------------------+-------------------+--------------------+---------------+---------------------+--------------------+----------------+-------------+----------------+--------------------+------------------+---------------------+-----------------------+---------+---------+---------+---------+-------
 users       |  id          | bigint                      | int8                | NO          |                                        |                          |                64 |                    |             0 |                     |                    |                | YES         |      100       |          1         |          1       |    BY DEFAULT       |                       |         | b       |         |         |    20
 users       |  rank        | integer                     | int4                | YES         |                                        |                          |                32 |                    |             0 |                     |                    |                | NO          |                |                    |                  |                     |                       | rank    | b       |         |         |    23
 users       |  c1          | smallint                    | int2                | NO          |           1000                         |                          |                16 |                    |             0 |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    21
 users       |  c2          | bit                         | bit                 | NO          |                                        |                        1 |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1560
 users       |  c3          | bit varying                 | varbit              | NO          |                                        |                       10 |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1562
 users       |  c4          | boolean                     | bool                | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    16
 users       |  c5          | bytea                       | bytea               | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    17
 users       |  c6          | character                   | bpchar              | NO          |                                        |                      100 |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1042
 users       |  c7          | character varying           | varchar             | NO          | 'logged_in'::character varying         |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1043
 users       |  c8          | cidr                        | cidr                | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   650
 users       |  c9          | circle                      | circle              | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   718
 users       |  c10         | date                        | date                | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1082
 users       |  c11         | time with time zone         | timetz              | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1266
 users       |  c12         | double precision            | float8              | NO          |                                        |                          |                53 |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   701
 users       |  c13         | real                        | float4              | NO          |           random()                     |                          |                24 |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   700
 users       |  c14         | json                        | json                | NO          |           '{}'::json                   |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   114
 users       |  c15         | jsonb                       | jsonb               | NO          |           '{}'::jsonb                  |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  3802
 users       |  c16         | money                       | money               | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   790
 users       |  c17         | numeric                     | numeric             | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1700
 users       |  c18         | numeric                     | numeric             | NO          |                                        |                          |                 4 |                    |             4 |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1700
 users       |  c19         | integer                     | int4                | NO          | nextval('t1_c19_seq'::regclass)        |                          |                32 |                    |             0 |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    23
 users       |  c20         | uuid                        | uuid                | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  2950
 users       |  c21         | xml                         | xml                 | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |   142
 users       |  c22         | ARRAY                       | integer[]           | YES         |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1007
 users       |  c23         | USER-DEFINED                | ltree               | YES         |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         | 16535
 users       |  c24         | USER-DEFINED                | state               | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | e       |         |         | 16774
 users       |  c25         | timestamp without time zone | timestamp           | NO          |            now()                       |                          |                   |                  4 |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1114
 users       |  c26         | timestamp with time zone    | timestamptz         | NO          |                                        |                          |                   |                  6 |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1184
 users       |  c27         | time without time zone      | time                | NO          |                                        |                          |                   |                  6 |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1266
 users       |  c28         | int                         | int8                | NO          |                                        |                          |                   |                  6 |               |                     |                    |                | NO          |                |                    |                  |                     |        (c1 + c2)      |         | b       |         |         |  1267
 users       |  c29         | interval                    | interval            | NO          |                                        |                          |                   |                  6 |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1268
 users       |  c30         | interval                    | interval            | NO          |                                        |                          |                   |                  6 |               |        MONTH        |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1269
 users       |  c31         | interval                    | interval            | NO          |                                        |                          |                   |                  6 |               | MINUTE TO SECOND(6) |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  1233
 users       |  c32         | bigint                      | int4                | NO          | nextval('public.t1_c32_seq'::regclass) |                          |                32 |                    |             0 |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    23
 users       |  c33         | USER-DEFINED                | test."status""."    | NO          |  'unknown'::test."status""."           |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | e       |         |         | 16775
 users       |  c34         | ARRAY                       | state[]             | NO          |                                        |                          |                   |                    |               |                     |                    |                | NO          |                |                    |                  |                     |                       |         | b       |  16774  |  e      | 16779
`))
				m.ExpectQuery(sqlutil.Escape(`SELECT enumtypid, enumlabel FROM pg_enum WHERE enumtypid IN ($1, $2)`)).
					WithArgs(16774, 16775).
					WillReturnRows(rows(`
 enumtypid | enumlabel
-----------+-----------
     16774 | on
     16774 | off
     16775 | unknown
`))
				m.noIndexes()
				m.noFKs()
				m.noChecks()
			},
			expect: func(require *require.Assertions, t *sqlschema.Table, err error) {
				p := func(i int) *int { return &i }
				require.NoError(err)
				require.Equal("users", t.Name)
				require.EqualValues([]*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Raw: "bigint", Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{Generation: "BY DEFAULT", Sequence: &Sequence{Start: 100, Increment: 1, Last: 1}}}},
					{Name: "rank", Type: &sqlschema.ColumnType{Raw: "integer", Nullable: true, Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "rank"}}},
					{Name: "c1", Type: &sqlschema.ColumnType{Raw: "smallint", Type: &sqlschema.IntegerType{T: "smallint"}}, Default: &sqlschema.Literal{V: "1000"}},
					{Name: "c2", Type: &sqlschema.ColumnType{Raw: "bit", Type: &BitType{T: "bit", Size: 1}}},
					{Name: "c3", Type: &sqlschema.ColumnType{Raw: "bit varying", Type: &BitType{T: "bit varying", Size: 10}}},
					{Name: "c4", Type: &sqlschema.ColumnType{Raw: "boolean", Type: &sqlschema.BoolType{T: "boolean"}}},
					{Name: "c5", Type: &sqlschema.ColumnType{Raw: "bytea", Type: &sqlschema.BinaryType{T: "bytea"}}},
					{Name: "c6", Type: &sqlschema.ColumnType{Raw: "character", Type: &sqlschema.StringType{T: "character", Size: 100}}},
					{Name: "c7", Type: &sqlschema.ColumnType{Raw: "character varying", Type: &sqlschema.StringType{T: "character varying"}}, Default: &sqlschema.Literal{V: "'logged_in'"}},
					{Name: "c8", Type: &sqlschema.ColumnType{Raw: "cidr", Type: &NetworkType{T: "cidr"}}},
					{Name: "c9", Type: &sqlschema.ColumnType{Raw: "circle", Type: &sqlschema.SpatialType{T: "circle"}}},
					{Name: "c10", Type: &sqlschema.ColumnType{Raw: "date", Type: &sqlschema.TimeType{T: "date"}}},
					{Name: "c11", Type: &sqlschema.ColumnType{Raw: "time with time zone", Type: &sqlschema.TimeType{T: "time with time zone", Precision: p(0)}}},
					{Name: "c12", Type: &sqlschema.ColumnType{Raw: "double precision", Type: &sqlschema.FloatType{T: "double precision", Precision: 53}}},
					{Name: "c13", Type: &sqlschema.ColumnType{Raw: "real", Type: &sqlschema.FloatType{T: "real", Precision: 24}}, Default: &sqlschema.RawExpr{X: "random()"}},
					{Name: "c14", Type: &sqlschema.ColumnType{Raw: "json", Type: &sqlschema.JSONType{T: "json"}}, Default: &sqlschema.Literal{V: "'{}'"}},
					{Name: "c15", Type: &sqlschema.ColumnType{Raw: "jsonb", Type: &sqlschema.JSONType{T: "jsonb"}}, Default: &sqlschema.Literal{V: "'{}'"}},
					{Name: "c16", Type: &sqlschema.ColumnType{Raw: "money", Type: &CurrencyType{T: "money"}}},
					{Name: "c17", Type: &sqlschema.ColumnType{Raw: "numeric", Type: &sqlschema.DecimalType{T: "numeric"}}},
					{Name: "c18", Type: &sqlschema.ColumnType{Raw: "numeric", Type: &sqlschema.DecimalType{T: "numeric", Precision: 4, Scale: 4}}},
					{Name: "c19", Type: &sqlschema.ColumnType{Raw: "serial", Type: &SerialType{T: "serial", SequenceName: "t1_c19_seq"}}},
					{Name: "c20", Type: &sqlschema.ColumnType{Raw: "uuid", Type: &UUIDType{T: "uuid"}}},
					{Name: "c21", Type: &sqlschema.ColumnType{Raw: "xml", Type: &XMLType{T: "xml"}}},
					{Name: "c22", Type: &sqlschema.ColumnType{Raw: "ARRAY", Nullable: true, Type: &ArrayType{Type: &sqlschema.IntegerType{T: "integer"}, T: "integer[]"}}},
					{Name: "c23", Type: &sqlschema.ColumnType{Raw: "USER-DEFINED", Nullable: true, Type: &UserDefinedType{T: "ltree"}}},
					{Name: "c24", Type: &sqlschema.ColumnType{Raw: "state", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}, Schema: t.Schema}}},
					{Name: "c25", Type: &sqlschema.ColumnType{Raw: "timestamp without time zone", Type: &sqlschema.TimeType{T: "timestamp without time zone", Precision: p(4)}}, Default: &sqlschema.RawExpr{X: "now()"}},
					{Name: "c26", Type: &sqlschema.ColumnType{Raw: "timestamp with time zone", Type: &sqlschema.TimeType{T: "timestamp with time zone", Precision: p(6)}}},
					{Name: "c27", Type: &sqlschema.ColumnType{Raw: "time without time zone", Type: &sqlschema.TimeType{T: "time without time zone", Precision: p(6)}}},
					{Name: "c28", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}, Attrs: []sqlschema.Attr{&sqlschema.GeneratedExpr{Expr: "(c1 + c2)"}}},
					{Name: "c29", Type: &sqlschema.ColumnType{Raw: "interval", Type: &IntervalType{T: "interval", Precision: p(6)}}},
					{Name: "c30", Type: &sqlschema.ColumnType{Raw: "interval", Type: &IntervalType{T: "interval", F: "MONTH", Precision: p(6)}}},
					{Name: "c31", Type: &sqlschema.ColumnType{Raw: "interval", Type: &IntervalType{T: "interval", F: "MINUTE TO SECOND", Precision: p(6)}}},
					{Name: "c32", Type: &sqlschema.ColumnType{Raw: "bigserial", Type: &SerialType{T: "bigserial", SequenceName: "t1_c32_seq"}}},
					{Name: "c33", Type: &sqlschema.ColumnType{Raw: `status".`, Type: &sqlschema.EnumType{T: `status".`, Values: []string{"unknown"}, Schema: sqlschema.New("test")}}, Default: &sqlschema.Literal{V: "'unknown'"}},
					{Name: "c34", Type: &sqlschema.ColumnType{Raw: "ARRAY", Type: &ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}, Schema: t.Schema}}}},
				}, t.Columns)
			},
		},
		{
			name: "table indexes",
			before: func(m mock) {
				m.tableExists("public", "users", true)
				m.ExpectQuery(queryColumns).
					WithArgs("public", "users").
					WillReturnRows(rows(`
table_name | column_name |      data_type      | formatted |  is_nullable |         column_default          | character_maximum_length | numeric_precision | datetime_precision | numeric_scale | interval_type | character_set_name | collation_name | is_identity | identity_start | identity_increment |   identity_last  | identity_generation | generation_expression | comment | typtype | typelem | elemtyp |  oid
-----------+-------------+---------------------+-----------+--------------+---------------------------------+--------------------------+-------------------+--------------------+---------------+---------------+--------------------+----------------+-------------+----------------+--------------------+------------------+---------------------+-----------------------+---------+---------+---------+---------+-------
users      | id          | bigint              | int8      |  NO          |                                 |                          |                64 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    20
users      | c1          | smallint            | int2      |  NO          |                                 |                          |                16 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    21
users      | parent_id   | bigint              | int8      |  YES         |                                 |                          |                64 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    22
`))
				m.ExpectQuery(queryIndexes).
					WithArgs("public", "users").
					WillReturnRows(rows(`
   table_name   |    index_name   | index_type  | column_name | included | primary | unique | constraint_type | predicate             |   expression              | desc | nulls_first | nulls_last | comment   | options
----------------+-----------------+-------------+-------------+----------+---------+--------+-----------------+-----------------------+---------------------------+------+-------------+------------+-----------+-----------
users           | idx             | hash        |             | f        | f       | f      |                 |                       | "left"((c11)::text, 100)  | t    | t           | f          | boring    |
users           | idx1            | btree       |             | f        | f       | f      |                 | (id <> NULL::integer) | "left"((c11)::text, 100)  | t    | t           | f          |           |
users           | t1_c1_key       | btree       | c1          | f        | f       | t      | u               |                       | c1                        | t    | t           | f          |           |
users           | t1_pkey         | btree       | id          | f        | t       | t      | p               |                       | id                        | t    | f           | f          |           |
users           | idx4            | btree       | c1          | f        | f       | t      |                 |                       | c1                        | f    | f           | f          |           |
users           | idx4            | btree       | id          | f        | f       | t      |                 |                       | id                        | f    | f           | t          |           |
users           | idx5            | btree       | c1          | f        | f       | t      |                 |                       | c1                        | f    | f           | f          |           |
users           | idx5            | btree       |             | f        | f       | t      |                 |                       | coalesce(parent_id, 0)    | f    | f           | f          |           |
users           | idx6            | brin        | c1          | f        | f       | t      |                 |                       |                           | f    | f           | f          |           | {autosummarize=true,pages_per_range=2}
users           | idx2            | btree       |             | f        | f       | f      |                 |                       | ((c * 2))                 | f    | f           | t          |           |
users           | idx2            | btree       | c1          | f        | f       | f      |                 |                       | c                         | f    | f           | t          |           |
users           | idx2            | btree       | id          | f        | f       | f      |                 |                       | d                         | f    | f           | t          |           |
users           | idx2            | btree       | c1          | t        | f       | f      |                 |                       | c                         |      |             |            |           |
users           | idx2            | btree       | parent_id   | t        | f       | f      |                 |                       | d                         |      |             |            |           |
`))
				m.noFKs()
				m.noChecks()
			},
			expect: func(require *require.Assertions, t *sqlschema.Table, err error) {
				require.NoError(err)
				require.Equal("users", t.Name)
				columns := []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Raw: "bigint", Type: &sqlschema.IntegerType{T: "bigint"}}},
					{Name: "c1", Type: &sqlschema.ColumnType{Raw: "smallint", Type: &sqlschema.IntegerType{T: "smallint"}}},
					{Name: "parent_id", Type: &sqlschema.ColumnType{Raw: "bigint", Nullable: true, Type: &sqlschema.IntegerType{T: "bigint"}}},
				}
				indexes := []*sqlschema.Index{
					{Name: "idx", Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "hash"}, &sqlschema.Comment{Text: "boring"}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Expr: &sqlschema.RawExpr{X: `"left"((c11)::text, 100)`}, Descending: true, Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsFirst: true}}}}},
					{Name: "idx1", Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "btree"}, &IndexPredicate{Predicate: `(id <> NULL::integer)`}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Expr: &sqlschema.RawExpr{X: `"left"((c11)::text, 100)`}, Descending: true, Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsFirst: true}}}}},
					{Name: "t1_c1_key", Unique: true, Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "btree"}, &ConstraintType{T: "u"}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: columns[1], Descending: true, Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsFirst: true}}}}},
					{Name: "idx4", Unique: true, Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "btree"}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: columns[1]}, {Seq: 2, Column: columns[0], Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsLast: true}}}}},
					{Name: "idx5", Unique: true, Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "btree"}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: columns[1]}, {Seq: 2, Expr: &sqlschema.RawExpr{X: `coalesce(parent_id, 0)`}}}},
					{Name: "idx6", Unique: true, Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "brin"}, &IndexStorageParams{AutoSummarize: true, PagesPerRange: 2}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: columns[1]}}},
					{Name: "idx2", Unique: false, Table: t, Attrs: []sqlschema.Attr{&IndexType{T: "btree"}, &IndexInclude{Columns: columnNames(columns[1:])}}, Parts: []*sqlschema.IndexPart{{Seq: 1, Expr: &sqlschema.RawExpr{X: `((c * 2))`}, Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsLast: true}}}, {Seq: 2, Column: columns[1], Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsLast: true}}}, {Seq: 3, Column: columns[0], Attrs: []sqlschema.Attr{&IndexColumnProperty{NullsLast: true}}}}},
				}
				pk := &sqlschema.Index{
					Name:   "t1_pkey",
					Unique: true,
					Table:  t,
					Attrs:  []sqlschema.Attr{&IndexType{T: "btree"}, &ConstraintType{T: "p"}},
					Parts:  []*sqlschema.IndexPart{{Seq: 1, Column: columns[0], Descending: true}},
				}
				columns[0].Indexes = append(columns[0].Indexes, pk, indexes[3], indexes[6])
				columns[1].Indexes = indexes[2:]
				require.EqualValues(columns, t.Columns)
				require.EqualValues(indexes, t.Indexes)
				require.EqualValues(pk, t.PrimaryKey)
			},
		},
		{
			name: "fks",
			before: func(m mock) {
				m.tableExists("public", "users", true)
				m.ExpectQuery(queryColumns).
					WithArgs("public", "users").
					WillReturnRows(rows(`
table_name | column_name |      data_type      | formatted | is_nullable |         column_default          | character_maximum_length | numeric_precision | datetime_precision | numeric_scale | interval_type | character_set_name | collation_name | is_identity | identity_start | identity_increment |   identity_last  | identity_generation | generation_expression | comment | typtype | typelem | elemtyp |  oid
-----------+-------------+---------------------+-----------+-------------+---------------------------------+--------------------------+-------------------+--------------------+---------------+---------------+--------------------+----------------+-------------+----------------+--------------------+------------------+---------------------+-----------------------+---------+---------+---------+---------+-------
users      | id          | integer             | int       | NO          |                                 |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    20
users      | oid         | integer             | int       | NO          |                                 |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    21
users      | uid         | integer             | int       | NO          |                                 |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |    21
`))
				m.noIndexes()
				m.ExpectQuery(queryFKs).
					WithArgs("public", "users").
					WillReturnRows(rows(`
constraint_name | table_name | column_name | table_schema | referenced_table_name | referenced_column_name | referenced_schema_name | update_rule | delete_rule
-----------------+------------+-------------+--------------+-----------------------+------------------------+------------------------+-------------+-------------
multi_column    | users      | id          | public       | t1                    | gid                    | public                 | NO ACTION   | CASCADE
multi_column    | users      | id          | public       | t1                    | xid                    | public                 | NO ACTION   | CASCADE
multi_column    | users      | oid         | public       | t1                    | gid                    | public                 | NO ACTION   | CASCADE
multi_column    | users      | oid         | public       | t1                    | xid                    | public                 | NO ACTION   | CASCADE
self_reference  | users      | uid         | public       | users                 | id                     | public                 | NO ACTION   | CASCADE
`))
				m.noChecks()
			},
			expect: func(require *require.Assertions, t *sqlschema.Table, err error) {
				require.NoError(err)
				require.Equal("users", t.Name)
				require.Equal("public", t.Schema.Name)
				fks := []*sqlschema.ForeignKey{
					{Name: "multi_column", Table: t, OnUpdate: sqlschema.NoAction, OnDelete: sqlschema.Cascade, RefTable: &sqlschema.Table{Name: "t1", Schema: t.Schema}, RefColumns: []*sqlschema.Column{{Name: "gid"}, {Name: "xid"}}},
					{Name: "self_reference", Table: t, OnUpdate: sqlschema.NoAction, OnDelete: sqlschema.Cascade, RefTable: t},
				}
				columns := []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}, ForeignKeys: fks[0:1]},
					{Name: "oid", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}, ForeignKeys: fks[0:1]},
					{Name: "uid", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}, ForeignKeys: fks[1:2]},
				}
				fks[0].Columns = columns[:2]
				fks[1].Columns = columns[2:]
				fks[1].RefColumns = columns[:1]
				require.EqualValues(columns, t.Columns)
				require.EqualValues(fks, t.ForeignKeys)
			},
		},
		{
			name: "check",
			before: func(m mock) {
				m.tableExists("public", "users", true)
				m.ExpectQuery(queryColumns).
					WithArgs("public", "users").
					WillReturnRows(rows(`
table_name |column_name | data_type | formatted | is_nullable | column_default | character_maximum_length | numeric_precision | datetime_precision | numeric_scale | interval_type | character_set_name | collation_name | is_identity | identity_start | identity_increment |   identity_last  | identity_generation | generation_expression | comment | typtype | typelem | elemtyp | oid
-----------+------------+-----------+-----------+-------------+----------------+--------------------------+-------------------+--------------------+---------------+---------------+--------------------+----------------+-------------+----------------+--------------------+------------------+---------------------+-----------------------+---------+---------+---------+---------+-----
users      | c1         | integer   | int4      | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
users      | c2         | integer   | int4      | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
users      | c3         | integer   | int4      | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
`))
				m.noIndexes()
				m.noFKs()
				m.ExpectQuery(queryChecks).
					WithArgs("public", "users").
					WillReturnRows(rows(`
table_name   | constraint_name    |       expression        | column_name | column_indexes | no_inherit
-------------+--------------------+-------------------------+-------------+----------------+----------------
users        | boring             | (c1 > 1)                | c1          | {1}            | t
users        | users_c2_check     | (c2 > 0)                | c2          | {2}            | f
users        | users_c2_check1    | (c2 > 0)                | c2          | {2}            | f
users        | users_check        | ((c2 + c1) > 2)         | c2          | {2,1}          | f
users        | users_check        | ((c2 + c1) > 2)         | c1          | {2,1}          | f
users        | users_check1       | (((c2 + c1) + c3) > 10) | c2          | {2,1,3}        | f
users        | users_check1       | (((c2 + c1) + c3) > 10) | c1          | {2,1,3}        | f
users        | users_check1       | (((c2 + c1) + c3) > 10) | c3          | {2,1,3}        | f
`))
				m.noChecks()
			},
			expect: func(require *require.Assertions, t *sqlschema.Table, err error) {
				require.NoError(err)
				require.Equal("users", t.Name)
				require.Equal("public", t.Schema.Name)
				require.EqualValues([]*sqlschema.Column{
					{Name: "c1", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}},
					{Name: "c2", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}},
					{Name: "c3", Type: &sqlschema.ColumnType{Raw: "integer", Type: &sqlschema.IntegerType{T: "integer"}}},
				}, t.Columns)
				require.EqualValues([]sqlschema.Attr{
					&sqlschema.Check{Name: "boring", Expr: "(c1 > 1)", Attrs: []sqlschema.Attr{&CheckColumns{Columns: []string{"c1"}}, &NoInherit{}}},
					&sqlschema.Check{Name: "users_c2_check", Expr: "(c2 > 0)", Attrs: []sqlschema.Attr{&CheckColumns{Columns: []string{"c2"}}}},
					&sqlschema.Check{Name: "users_c2_check1", Expr: "(c2 > 0)", Attrs: []sqlschema.Attr{&CheckColumns{Columns: []string{"c2"}}}},
					&sqlschema.Check{Name: "users_check", Expr: "((c2 + c1) > 2)", Attrs: []sqlschema.Attr{&CheckColumns{Columns: []string{"c2", "c1"}}}},
					&sqlschema.Check{Name: "users_check1", Expr: "(((c2 + c1) + c3) > 10)", Attrs: []sqlschema.Attr{&CheckColumns{Columns: []string{"c2", "c1", "c3"}}}},
				}, t.Attrs)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, m, err := sqlmock.New()
			require.NoError(t, err)
			mk := mock{m}
			mk.version("130000")
			var drv sqlschema.Driver
			drv, err = Open(db)
			require.NoError(t, err)
			mk.ExpectQuery(sqlutil.Escape(fmt.Sprintf(schemasQueryArgs, "= $1"))).
				WithArgs("public").
				WillReturnRows(rows(`
    schema_name
--------------------
 public
`))
			tt.before(mk)
			s, err := drv.InspectSchema(context.Background(), "public", nil)
			require.NoError(t, err)
			tt.expect(require.New(t), s.Tables[0], err)
		})
	}
}

func TestDriver_InspectPartitionedTable(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mk := mock{m}
	mk.version("130000")
	drv, err := Open(db)
	require.NoError(t, err)
	mk.ExpectQuery(sqlutil.Escape(fmt.Sprintf(schemasQueryArgs, "= CURRENT_SCHEMA()"))).
		WillReturnRows(rows(`
   schema_name
--------------------
public
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1"))).
		WithArgs("public").
		WillReturnRows(rows(`
 table_schema | table_name  | comment | partition_attrs | partition_strategy |                  partition_exprs
--------------+-------------+---------+-----------------+--------------------+----------------------------------------------------
 public       | logs1       |         |                 |                    |
 public       | logs2       |         | 1               | r                  |
 public       | logs3       |         | 2 0 0           | l                  | (a + b), (a + (b * 2))

`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(columnsQuery, "$2, $3, $4"))).
		WithArgs("public", "logs1", "logs2", "logs3").
		WillReturnRows(rows(`
table_name |column_name | data_type | formatted | is_nullable | column_default | character_maximum_length | numeric_precision | datetime_precision | numeric_scale | interval_type | character_set_name | collation_name | is_identity | identity_start | identity_increment |   identity_last  | identity_generation | generation_expression | comment | typtype | typelem | elemtyp | oid
-----------+------------+-----------+-----------+-------------+----------------+--------------------------+-------------------+--------------------+---------------+---------------+--------------------+----------------+-------------+----------------+--------------------+------------------+---------------------+-----------------------+---------+---------+---------+---------+-----
logs1      | c1         | integer   | integer   | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
logs2      | c2         | integer   | integer   | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
logs2      | c3         | integer   | integer   | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
logs3      | c4         | integer   | integer   | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
logs3      | c5         | integer   | integer   | NO          |                |                          |                32 |                    |             0 |               |                    |                | NO          |                |                    |                  |                     |                       |         | b       |         |         |  23
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(indexesQuery, "$2, $3, $4"))).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "column_name", "primary", "unique", "constraint_type", "predicate", "expression"}))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(fksQuery, "$2, $3, $4"))).
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "table_name", "column_name", "referenced_table_name", "referenced_column_name", "referenced_table_schema", "update_rule", "delete_rule"}))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(checksQuery, "$2, $3, $4"))).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "constraint_name", "expression", "column_name", "column_indexes"}))
	s, err := drv.InspectSchema(context.Background(), "", &sqlschema.InspectOptions{})
	require.NoError(t, err)

	t1, ok := s.Table("logs1")
	require.True(t, ok)
	require.Empty(t, t1.Attrs)

	t2, ok := s.Table("logs2")
	require.True(t, ok)
	require.Len(t, t2.Attrs, 1)
	key := t2.Attrs[0].(*Partition)
	require.Equal(t, PartitionTypeRange, key.T)
	require.Equal(t, []*PartitionPart{
		{Column: "c2"},
	}, key.Parts)

	t3, ok := s.Table("logs3")
	require.True(t, ok)
	require.Len(t, t3.Attrs, 1)
	key = t3.Attrs[0].(*Partition)
	require.Equal(t, PartitionTypeList, key.T)
	require.Equal(t, []*PartitionPart{
		{Column: "c5"},
		{Expr: &sqlschema.RawExpr{X: "(a + b)"}},
		{Expr: &sqlschema.RawExpr{X: "(a + (b * 2))"}},
	}, key.Parts)
}

func TestDriver_InspectSchema(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mk := mock{m}
	mk.version("130000")
	drv, err := Open(db)
	require.NoError(t, err)
	mk.ExpectQuery(sqlutil.Escape(fmt.Sprintf(schemasQueryArgs, "= CURRENT_SCHEMA()"))).
		WillReturnRows(rows(`
   schema_name
--------------------
test
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1"))).
		WithArgs("test").
		WillReturnRows(sqlmock.NewRows([]string{"table_schema", "table_name", "comment", "partition_attrs", "partition_strategy", "partition_exprs"}))
	s, err := drv.InspectSchema(context.Background(), "", &sqlschema.InspectOptions{})
	require.NoError(t, err)
	require.EqualValues(t, func() *sqlschema.Schema {
		r := &sqlschema.Realm{
			Schemas: []*sqlschema.Schema{
				{
					Name: "test",
				},
			},
			// Server default configuration.
			Attrs: []sqlschema.Attr{
				&sqlschema.Collation{
					V: "en_US.utf8",
				},
				&CType{
					V: "en_US.utf8",
				},
			},
		}
		r.Schemas[0].Realm = r
		return r.Schemas[0]
	}(), s)
}

func TestDriver_Realm(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mk := mock{m}
	mk.version("130000")
	drv, err := Open(db)
	require.NoError(t, err)
	mk.ExpectQuery(sqlutil.Escape(schemasQuery)).
		WillReturnRows(rows(`
   schema_name
--------------------
test
public
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1, $2"))).
		WithArgs("test", "public").
		WillReturnRows(sqlmock.NewRows([]string{"table_schema", "table_name", "comment", "partition_attrs", "partition_strategy", "partition_exprs"}))
	realm, err := drv.InspectRealm(context.Background(), &sqlschema.InspectRealmOption{})
	require.NoError(t, err)
	require.EqualValues(t, func() *sqlschema.Realm {
		r := &sqlschema.Realm{
			Schemas: []*sqlschema.Schema{
				{
					Name: "test",
				},
				{
					Name: "public",
				},
			},
			// Server default configuration.
			Attrs: []sqlschema.Attr{
				&sqlschema.Collation{
					V: "en_US.utf8",
				},
				&CType{
					V: "en_US.utf8",
				},
			},
		}
		r.Schemas[0].Realm = r
		r.Schemas[1].Realm = r
		return r
	}(), realm)

	mk.ExpectQuery(sqlutil.Escape(fmt.Sprintf(schemasQueryArgs, "IN ($1, $2)"))).
		WithArgs("test", "public").
		WillReturnRows(rows(`
   schema_name
--------------------
  test
  public
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1, $2"))).
		WithArgs("test", "public").
		WillReturnRows(sqlmock.NewRows([]string{"table_schema", "table_name", "comment", "partition_attrs", "partition_strategy", "partition_exprs"}))
	realm, err = drv.InspectRealm(context.Background(), &sqlschema.InspectRealmOption{Schemas: []string{"test", "public"}})
	require.NoError(t, err)
	require.EqualValues(t, func() *sqlschema.Realm {
		r := &sqlschema.Realm{
			Schemas: []*sqlschema.Schema{
				{
					Name: "test",
				},
				{
					Name: "public",
				},
			},
			// Server default configuration.
			Attrs: []sqlschema.Attr{
				&sqlschema.Collation{
					V: "en_US.utf8",
				},
				&CType{
					V: "en_US.utf8",
				},
			},
		}
		r.Schemas[0].Realm = r
		r.Schemas[1].Realm = r
		return r
	}(), realm)

	mk.ExpectQuery(sqlutil.Escape(fmt.Sprintf(schemasQueryArgs, "= $1"))).
		WithArgs("test").
		WillReturnRows(rows(`
 schema_name
--------------------
 test
`))
	m.ExpectQuery(sqlutil.Escape(fmt.Sprintf(tablesQuery, "$1"))).
		WithArgs("test").
		WillReturnRows(sqlmock.NewRows([]string{"table_schema", "table_name", "comment", "partition_attrs", "partition_strategy", "partition_exprs"}))
	realm, err = drv.InspectRealm(context.Background(), &sqlschema.InspectRealmOption{Schemas: []string{"test"}})
	require.NoError(t, err)
	require.EqualValues(t, func() *sqlschema.Realm {
		r := &sqlschema.Realm{
			Schemas: []*sqlschema.Schema{
				{
					Name: "test",
				},
			},
			// Server default configuration.
			Attrs: []sqlschema.Attr{
				&sqlschema.Collation{
					V: "en_US.utf8",
				},
				&CType{
					V: "en_US.utf8",
				},
			},
		}
		r.Schemas[0].Realm = r
		return r
	}(), realm)
}

func TestInspectMode_InspectRealm(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mk := mock{m}
	mk.version("130000")
	mk.ExpectQuery(sqlutil.Escape(schemasQuery)).
		WillReturnRows(rows(`
   schema_name
--------------------
test
public
`))
	drv, err := Open(db)
	require.NoError(t, err)
	realm, err := drv.InspectRealm(context.Background(), &sqlschema.InspectRealmOption{Mode: sqlschema.InspectSchemas})
	require.NoError(t, err)
	require.EqualValues(t, func() *sqlschema.Realm {
		r := &sqlschema.Realm{
			Schemas: []*sqlschema.Schema{
				{
					Name: "test",
				},
				{
					Name: "public",
				},
			},
			// Server default configuration.
			Attrs: []sqlschema.Attr{
				&sqlschema.Collation{
					V: "en_US.utf8",
				},
				&CType{
					V: "en_US.utf8",
				},
			},
		}
		r.Schemas[0].Realm = r
		r.Schemas[1].Realm = r
		return r
	}(), realm)
}

type mock struct {
	sqlmock.Sqlmock
}

func (m mock) version(version string) {
	m.ExpectQuery(sqlutil.Escape(paramsQuery)).
		WillReturnRows(rows(`
  setting
------------
 ` + version + `
 en_US.utf8
 en_US.utf8
`))
}

func (m mock) tableExists(schema, table string, exists bool) {
	rows := sqlmock.NewRows([]string{"table_schema", "table_name", "table_comment", "partition_attrs", "partition_strategy", "partition_exprs"})
	if exists {
		rows.AddRow(schema, table, nil, nil, nil, nil)
	}
	m.ExpectQuery(queryTables).
		WithArgs(schema).
		WillReturnRows(rows)
}

func (m mock) noIndexes() {
	m.ExpectQuery(queryIndexes).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "index_name", "column_name", "primary", "unique", "constraint_type", "predicate", "expression", "options"}))
}

func (m mock) noFKs() {
	m.ExpectQuery(queryFKs).
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name", "table_name", "column_name", "referenced_table_name", "referenced_column_name", "referenced_table_schema", "update_rule", "delete_rule"}))
}

func (m mock) noChecks() {
	m.ExpectQuery(queryChecks).
		WillReturnRows(sqlmock.NewRows([]string{"table_name", "constraint_name", "expression", "column_name", "column_indexes"}))
}
