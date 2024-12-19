package pggenerate_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/node/engine/parse"
	pggenerate "github.com/kwilteam/kwil-db/node/engine/pg_generate"
	"github.com/stretchr/testify/require"
)

func Test_PgGenerate(t *testing.T) {
	type testcase struct {
		name    string
		sql     string
		want    string
		params  []string
		wantErr bool
	}

	tests := []testcase{
		{
			name:   "Simple Insert with two params",
			sql:    "INSERT INTO tbl VALUES ($a, $b);",
			want:   "INSERT INTO kwil.tbl VALUES ($1, $2);",
			params: []string{"$a", "$b"},
		},
		{
			name:   "select with @caller",
			sql:    "SELECT * FROM tbl WHERE col = @caller;",
			want:   "SELECT * FROM tbl WHERE col = $1;",
			params: []string{"@caller"},
		},
		{
			name:   "Insert with named columns and params",
			sql:    "INSERT INTO tbl (col1, col2) VALUES ($foo, $bar);",
			want:   "INSERT INTO kwil.tbl (col1, col2) VALUES ($1, $2);",
			params: []string{"$foo", "$bar"},
		},
		{
			name:   "Update statement",
			sql:    "UPDATE tbl SET col1 = $x, col2 = $y WHERE col3 = $z;",
			want:   "UPDATE kwil.tbl SET col1 = $1, col2 = $2 WHERE col3 = $3;",
			params: []string{"$x", "$y", "$z"},
		},
		{
			name:   "Select with one param",
			sql:    "SELECT * FROM tbl WHERE col = $param;",
			want:   "SELECT * FROM tbl WHERE col = $1;",
			params: []string{"$param"},
		},
		{
			name:   "Delete with a param",
			sql:    "DELETE FROM tbl WHERE id = $id;",
			want:   "DELETE FROM kwil.tbl WHERE id = $1;",
			params: []string{"$id"},
		},
		{
			name:   "Complex select with multiple params",
			sql:    "SELECT col1, col2 FROM tbl WHERE col1 = $foo AND col2 IN ($bar, $baz);",
			want:   "SELECT col1, col2 FROM tbl WHERE col1 = $1 AND col2 IN ($2, $3);",
			params: []string{"$foo", "$bar", "$baz"},
		},
		{
			name:   "Repeated parameter name",
			sql:    "SELECT * FROM tbl WHERE col1 = $foo AND col2 = $foo;",
			want:   "SELECT * FROM tbl WHERE col1 = $1 AND col2 = $1;",
			params: []string{"$foo"},
		},
		{
			name:   "Mixed case parameter name",
			sql:    "SELECT * FROM tbl WHERE UserId = $UserId;",
			want:   "SELECT * FROM tbl WHERE userid = $1;",
			params: []string{"$userid"},
		},
		{
			name:   "Parameter name with underscore",
			sql:    "SELECT * FROM tbl WHERE col = $some_param_name;",
			want:   "SELECT * FROM tbl WHERE col = $1;",
			params: []string{"$some_param_name"},
		},
		{
			name:   "Multiple parameters used in multiple places",
			sql:    "UPDATE tbl SET col1 = $a, col2 = $b WHERE col3 = $a;",
			want:   "UPDATE kwil.tbl SET col1 = $1, col2 = $2 WHERE col3 = $1;",
			params: []string{"$a", "$b"},
		},
		{
			name: "No parameters",
			sql:  "SELECT * FROM tbl;",
			want: "SELECT * FROM tbl;",
		},
		{
			name:   "Parameter in function call",
			sql:    "SELECT * FROM tbl WHERE col = abs($pwd);",
			want:   "SELECT * FROM tbl WHERE col = abs($1);",
			params: []string{"$pwd"},
		},
		{
			name:   "Parameter in JOIN condition",
			sql:    "SELECT t1.col, t2.col FROM t1 JOIN t2 ON t1.id = t2.id AND t1.name = $name;",
			want:   "SELECT t1.col, t2.col FROM t1 INNER JOIN t2 ON t1.id = t2.id AND t1.name = $1;",
			params: []string{"$name"},
		},
		{
			name: "window function",
			sql:  "SELECT col1, col2, SUM(col3) OVER (PARTITION BY col1 ORDER BY col2) FROM tbl;",
			want: "SELECT col1, col2, sum(col3) OVER (PARTITION BY col1 ORDER BY col2 ASC NULLS LAST) FROM tbl;",
		},
		{
			name: "array access",
			sql:  "SELECT col1[1], col2[2] FROM tbl;",
			want: "SELECT col1[1], col2[2] FROM tbl;",
		},
		{
			name: "array slice",
			sql:  "SELECT col1[1:2], col2[2:], col3[:3] FROM tbl;",
			want: "SELECT col1[1:2], col2[2:], col3[:3] FROM tbl;",
		},
		{
			name: "make array",
			sql:  "SELECT ARRAY[col1, col2] FROM tbl;",
			want: "SELECT ARRAY[col1, col2] FROM tbl;",
		},
		{
			name: "type cast",
			sql:  "SELECT col1::INT8, (col2::TEXT)::INT8 FROM tbl;",
			want: "SELECT col1::INT8, (col2::TEXT)::INT8 FROM tbl;",
		},
		{
			name: "arithmetics",
			sql:  "SELECT col1 + col2, col1 - col2, col1 * col2, col1 / col2 FROM tbl;",
			want: "SELECT col1 + col2, col1 - col2, col1 * col2, col1 / col2 FROM tbl;",
		},
		{
			name: "comparison",
			sql:  "SELECT col1 = col2, col1 <> col2, col1 < col2, col1 <= col2, col1 > col2, col1 >= col2 FROM tbl;",
			want: "SELECT col1 = col2, col1 <> col2, col1 < col2, col1 <= col2, col1 > col2, col1 >= col2 FROM tbl;",
		},
		{
			name: "unary",
			sql:  "SELECT +col1, -col2 FROM tbl;",
			want: "SELECT +col1, -col2 FROM tbl;",
		},
		{
			name: "logical",
			sql:  "SELECT col1 AND col2, col1 OR col2, NOT col1 FROM tbl;",
			want: "SELECT col1 AND col2, col1 OR col2, NOT col1 FROM tbl;",
		},
		{
			name: "case",
			sql:  "SELECT CASE WHEN col1 = 1 THEN 'one' ELSE 'other' END FROM tbl;",
			want: "SELECT CASE WHEN col1 = 1 THEN 'one' ELSE 'other' END FROM tbl;",
		},
		{
			name: "collate",
			sql:  "SELECT col1 from tbl where name = 'foo' collate nocase;",
			want: "SELECT col1 FROM tbl WHERE name = 'foo' COLLATE nocase;",
		},
		{
			name: "is null",
			sql:  "SELECT col1 IS NULL, col2 IS NOT NULL FROM tbl;",
			want: "SELECT col1 IS NULL, col2 IS NOT NULL FROM tbl;",
		},
		{
			name: "between",
			sql:  "SELECT col1 BETWEEN 1 AND 10 FROM tbl;",
			want: "SELECT col1 BETWEEN 1 AND 10 FROM tbl;",
		},
		{
			name: "in",
			sql:  "SELECT col1 IN (1, 2, 3) FROM tbl;",
			want: "SELECT col1 IN (1, 2, 3) FROM tbl;",
		},
		{
			name: "like",
			sql:  "SELECT col1 LIKE 'foo%' FROM tbl WHERE col2 NOT LIKE '%bar' AND col3 ILIKE 'baz%';",
			want: "SELECT col1 LIKE 'foo%' FROM tbl WHERE col2 NOT LIKE '%bar' AND col3 ILIKE 'baz%';",
		},
		{
			name: "exists",
			sql:  "SELECT EXISTS (SELECT 1 FROM tbl WHERE col1 = 1);",
			want: "SELECT EXISTS (SELECT 1 FROM tbl WHERE col1 = 1);",
		},
		{
			name: "subquery",
			sql:  "SELECT (SELECT col1 FROM tbl WHERE col2 = 1) FROM tbl2;",
			want: "SELECT (SELECT col1 FROM tbl WHERE col2 = 1) FROM tbl2;",
		},
		{
			name: "common table expression",
			sql:  "WITH cte AS (SELECT * FROM tbl) SELECT * FROM cte;",
			want: "WITH cte AS (SELECT * FROM tbl) SELECT * FROM cte;",
		},
		{
			name: "recursive common table expression",
			sql:  "WITH RECURSIVE cte AS (SELECT * FROM tbl) SELECT * FROM cte;",
			want: "WITH RECURSIVE cte AS (SELECT * FROM tbl) SELECT * FROM cte;",
		},
		// ddl
		{
			name: "Create table",
			sql: `CREATE TABLE departments (
    department_id   UUID,
    department_code TEXT NOT NULL,
    department_name TEXT NOT NULL,
    location_id     INT8 DEFAULT 1,
    created_at      INT8,
    PRIMARY KEY (department_id, department_code),
    UNIQUE (department_name),
    CHECK (department_name <> ''),
	FOREIGN KEY (location_id) REFERENCES locations(location_id)
);`,
			want: `CREATE TABLE kwil.departments (
				department_id   UUID,
				department_code TEXT NOT NULL,
				department_name TEXT NOT NULL,
				location_id     INT8 DEFAULT 1,
				created_at      INT8,
				PRIMARY KEY (department_id, department_code),
				UNIQUE (department_name),
				CHECK (department_name <> ''),
				FOREIGN KEY (location_id) REFERENCES kwil.locations(location_id)
			);`,
		},
		{
			name: "Create table if not exists",
			sql: `CREATE TABLE IF NOT EXISTS departments (
				department_id   UUID PRIMARY KEY
		);`,
			want: `CREATE TABLE IF NOT EXISTS kwil.departments (
				department_id   UUID PRIMARY KEY
			);`,
		},
		{
			name: "add column",
			sql:  `ALTER TABLE departments ADD COLUMN department_head UUID;`,
			want: `ALTER TABLE kwil.departments ADD COLUMN department_head UUID;`,
		},
		{
			name: "drop column",
			sql:  `ALTER TABLE departments DROP COLUMN department_head;`,
			want: `ALTER TABLE kwil.departments DROP COLUMN department_head;`,
		},
		{
			name: "rename column",
			sql:  `ALTER TABLE departments RENAME COLUMN department_head TO head_department;`,
			want: `ALTER TABLE kwil.departments RENAME COLUMN department_head TO head_department;`,
		},
		{
			name: "rename table",
			sql:  `ALTER TABLE departments RENAME TO division;`,
			want: `ALTER TABLE kwil.departments RENAME TO division;`,
		},
		{
			name: "add table constraint",
			sql:  `ALTER TABLE departments ADD PRIMARY KEY (department_id);`,
			want: `ALTER TABLE kwil.departments ADD PRIMARY KEY (department_id);`,
		},
		{
			name: "drop table constraint",
			sql:  `ALTER TABLE departments DROP CONSTRAINT department_id;`,
			want: `ALTER TABLE kwil.departments DROP CONSTRAINT department_id;`,
		},
		{
			name: "add column constraint",
			sql:  `ALTER TABLE departments ALTER COLUMN department_head SET NOT NULL;`,
			want: `ALTER TABLE kwil.departments ALTER COLUMN department_head SET NOT NULL;`,
		},
		{
			name: "drop column constraint",
			sql:  `ALTER TABLE departments ALTER COLUMN department_head DROP NOT NULL;`,
			want: `ALTER TABLE kwil.departments ALTER COLUMN department_head DROP NOT NULL;`,
		},
		{
			name: "drop table",
			sql:  `DROP TABLE departments;`,
			want: `DROP TABLE kwil.departments;`,
		},
		{
			name: "drop table if exists cascade",
			sql:  `DROP TABLE IF EXISTS departments CASCADE;`,
			want: `DROP TABLE IF EXISTS kwil.departments CASCADE;`,
		},
		{
			name: "create index",
			sql:  `CREATE INDEX IF NOT EXISTS idx_department_name_id ON departments (department_name, department_id);`,
			want: `CREATE INDEX IF NOT EXISTS idx_department_name_id ON kwil.departments (department_name, department_id);`,
		},
		{
			name: "drop index",
			sql:  `DROP INDEX IF EXISTS idx_department_name_id;`,
			want: `DROP INDEX IF EXISTS kwil.idx_department_name_id;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parse.Parse(tt.sql)
			require.NoError(t, err)
			require.Len(t, parsed, 1)

			got, ps, err := pggenerate.GenerateSQL(parsed[0], "kwil")
			if err != nil {
				if !tt.wantErr {
					require.NoError(t, err)
				}
				return
			} else {
				require.Equal(t, tt.wantErr, false)
			}

			require.Equal(t, removeWhitespace(tt.want), removeWhitespace(got))
			require.EqualValues(t, tt.params, ps)
		})
	}
}

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}
