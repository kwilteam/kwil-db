package generate_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/generate"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/kwilteam/kwil-db/parse/postgres"
)

func TestGenerateDDLStatement(t *testing.T) {
	tests := []struct {
		name    string
		sql     *parse.SQLStatement
		want    string
		wantErr bool
	}{ // those are the same as what are in internal.parse.parse_test.Test_SQL, with 'want' and 'sql' swapped
		{
			name: "create table",
			sql: &parse.SQLStatement{
				SQL: &parse.CreateTableStatement{
					Name: "users",
					Columns: []*parse.Column{
						{
							Name: "id",
							Type: types.IntType,
							Constraints: []parse.Constraint{
								&parse.ConstraintPrimaryKey{},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Constraints: []parse.Constraint{
								&parse.ConstraintCheck{
									Param: &parse.ExpressionComparison{
										Left: &parse.ExpressionFunctionCall{
											Name: "length",
											Args: []parse.Expression{
												&parse.ExpressionColumn{
													Table:  "",
													Column: "name",
												},
											},
										},
										Right: &parse.ExpressionLiteral{
											Type:  types.IntType,
											Value: int64(10),
										},
										Operator: parse.ComparisonOperatorGreaterThan,
									},
								},
							},
						},
						{
							Name: "address",
							Type: types.TextType,
							Constraints: []parse.Constraint{
								&parse.ConstraintNotNull{},
								&parse.ConstraintDefault{
									Value: &parse.ExpressionLiteral{
										Type:  types.TextType,
										Value: "usa",
									},
								},
							},
						},
						{
							Name: "email",
							Type: types.TextType,
							Constraints: []parse.Constraint{
								&parse.ConstraintNotNull{},
								&parse.ConstraintUnique{},
							},
						},
						{
							Name: "city_id",
							Type: types.IntType,
						},
						{
							Name: "group_id",
							Type: types.IntType,
							Constraints: []parse.Constraint{
								&parse.ConstraintForeignKey{
									RefTable:  "groups",
									RefColumn: "id",
									Ons:       []parse.ForeignKeyActionOn{parse.ON_DELETE},
									Dos:       []parse.ForeignKeyActionDo{parse.DO_CASCADE},
								},
							},
						},
					},
					Indexes: []*parse.Index{
						{
							Name:    "group_name_unique",
							Columns: []string{"group_id", "name"},
							Type:    parse.IndexTypeUnique,
						},
						{
							Name:    "ithome",
							Columns: []string{"name", "address"},
							Type:    parse.IndexTypeBTree,
						},
					},
					Constraints: []parse.Constraint{
						&parse.ConstraintForeignKey{
							Name:      "city_fk",
							RefTable:  "cities",
							RefColumn: "id",
							Column:    "city_id",
							Ons:       []parse.ForeignKeyActionOn{parse.ON_UPDATE},
							Dos:       []parse.ForeignKeyActionDo{parse.DO_NO_ACTION},
						},
						&parse.ConstraintCheck{
							Param: &parse.ExpressionComparison{
								Left: &parse.ExpressionFunctionCall{
									Name: "length",
									Args: []parse.Expression{
										&parse.ExpressionColumn{
											Table:  "",
											Column: "email",
										},
									},
								},
								Right: &parse.ExpressionLiteral{
									Type:  types.IntType,
									Value: int64(1),
								},
								Operator: parse.ComparisonOperatorGreaterThan,
							},
						},
						&parse.ConstraintUnique{
							Columns: []string{
								"city_id",
								"address",
							},
						},
					},
				},
			},
			want: `CREATE TABLE users (
  id int PRIMARY KEY,
  name text CHECK(length(name) > 10),
  address text NOT NULL DEFAULT 'usa',
  email text NOT NULL UNIQUE,
  city_id int,
  group_id int REFERENCES groups(id) ON DELETE CASCADE,
  CONSTRAINT city_fk FOREIGN KEY (city_id) REFERENCES cities(id) ON UPDATE NO ACTION,
  CHECK(length(email) > 1),
  UNIQUE (city_id, address),
  UNIQUE INDEX group_name_unique (group_id, name),
  INDEX ithome (name, address)
);`,
		},
		{
			name: "alter table add column constraint NOT NULL",
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.AddColumnConstraint{
						Column: "name",
						Type:   parse.NOT_NULL,
					},
				},
			},
			want: "ALTER TABLE user ALTER COLUMN name SET NOT NULL;",
		},
		{
			name: "alter table add column constraint DEFAULT",
			want: `ALTER TABLE user ALTER COLUMN name SET DEFAULT 10;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.AddColumnConstraint{
						Column: "name",
						Type:   parse.DEFAULT,
						Value: &parse.ExpressionLiteral{
							Type:  types.IntType,
							Value: int64(10),
						},
					},
				},
			},
		},
		{
			name: "alter table drop column constraint NOT NULL",
			want: `ALTER TABLE user ALTER COLUMN name DROP NOT NULL;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.DropColumnConstraint{
						Column: "name",
						Type:   parse.NOT_NULL,
					},
				},
			},
		},
		{
			name: "alter table drop column constraint DEFAULT",
			want: `ALTER TABLE user ALTER COLUMN name DROP DEFAULT;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.DropColumnConstraint{
						Column: "name",
						Type:   parse.DEFAULT,
					},
				},
			},
		},
		{
			name: "alter table drop column constraint named",
			want: `ALTER TABLE user ALTER COLUMN name DROP CONSTRAINT abc;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.DropColumnConstraint{
						Column: "name",
						Type:   parse.NAMED,
						Name:   "abc",
					},
				},
			},
		},
		{
			name: "alter table add column",
			want: `ALTER TABLE user ADD COLUMN abc int;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.AddColumn{
						Name: "abc",
						Type: types.IntType,
					},
				},
			},
		},
		{
			name: "alter table drop column",
			want: `ALTER TABLE user DROP COLUMN abc;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.DropColumn{
						Name: "abc",
					},
				},
			},
		},

		{
			name: "alter table rename column",
			want: `ALTER TABLE user RENAME COLUMN abc TO def;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.RenameColumn{
						OldName: "abc",
						NewName: "def",
					},
				},
			},
		},
		{
			name: "alter table rename table",
			want: `ALTER TABLE user RENAME TO account;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.RenameTable{
						Name: "account",
					},
				},
			},
		},
		{
			name: "alter table add constraint fk",
			want: `ALTER TABLE user ADD CONSTRAINT new_fk FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE CASCADE;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.AddTableConstraint{
						Cons: &parse.ConstraintForeignKey{
							Name:      "new_fk",
							RefTable:  "cities",
							RefColumn: "id",
							Column:    "city_id",
							Ons:       []parse.ForeignKeyActionOn{parse.ON_DELETE},
							Dos:       []parse.ForeignKeyActionDo{parse.DO_CASCADE},
						},
					},
				},
			},
		},
		{
			name: "alter table drop constraint",
			want: `ALTER TABLE user DROP CONSTRAINT abc;`,
			sql: &parse.SQLStatement{
				SQL: &parse.AlterTableStatement{
					Table: "user",
					Action: &parse.DropTableConstraint{
						Name: "abc",
					},
				},
			},
		},
		{
			name: "create index",
			want: `CREATE INDEX abc ON user(name);`,
			sql: &parse.SQLStatement{
				SQL: &parse.CreateIndexStatement{
					Index: parse.Index{
						Name:    "abc",
						On:      "user",
						Columns: []string{"name"},
						Type:    parse.IndexTypeBTree,
					},
				},
			},
		},
		{
			name: "create unique index",
			want: `CREATE UNIQUE INDEX abc ON user(name);`,
			sql: &parse.SQLStatement{
				SQL: &parse.CreateIndexStatement{
					Index: parse.Index{
						Name:    "abc",
						On:      "user",
						Columns: []string{"name"},
						Type:    parse.IndexTypeUnique,
					},
				},
			},
		},
		{
			name: "create index with no name",
			want: `CREATE INDEX ON user(name);`,
			sql: &parse.SQLStatement{
				SQL: &parse.CreateIndexStatement{
					Index: parse.Index{
						On:      "user",
						Columns: []string{"name"},
						Type:    parse.IndexTypeBTree,
					},
				},
			},
		},
		{
			name: "drop index",
			want: `DROP INDEX abc;`,
			sql: &parse.SQLStatement{
				SQL: &parse.DropIndexStatement{
					Name: "abc",
				},
			},
		},

		{
			name: "drop index check exist",
			want: `DROP INDEX IF EXISTS abc;`,
			sql: &parse.SQLStatement{
				SQL: &parse.DropIndexStatement{
					Name:       "abc",
					CheckExist: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generate.DDL(tt.sql)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestGenerateDDL(t *testing.T) {
	type args struct {
		table *types.Table
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "table with composite primary key",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					Indexes: []*types.Index{
						{
							Name:    "test_index",
							Type:    types.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
						{
							Name:    "CompositePrimaryKey",
							Type:    types.PRIMARY,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			want: []string{
				`CREATE TABLE "dbid"."test" ("id" INT8 NOT NULL, "name" TEXT NOT NULL DEFAULT 'foo', PRIMARY KEY ("id", "name"));`,
				`CREATE UNIQUE INDEX "test_index" ON "dbid"."test" ("id", "name");`,
			},
		},
		{
			name: "table with composite primary key and composite index",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					Indexes: []*types.Index{
						{
							Name:    "test_index",
							Type:    types.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "table with foreign key on update set cascade",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					ForeignKeys: []*types.ForeignKey{
						{

							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_CASCADE,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "dbid"."test" ("id" INT8, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "dbid"."users"("username") ON UPDATE CASCADE, PRIMARY KEY ("id"));`},
		},
		{
			name: "table with multiple foreign keys and multiple actions per foreign key",
			args: args{
				table: &types.Table{
					Name: "table1",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					ForeignKeys: []*types.ForeignKey{
						{
							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_CASCADE,
								},
								{
									On: types.ON_DELETE,
									Do: types.DO_SET_DEFAULT,
								},
							},
						},
						{
							ChildKeys:   []string{"id", "name"},
							ParentKeys:  []string{"id", "username"},
							ParentTable: "table2",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_SET_NULL,
								},
								{
									On: types.ON_DELETE,
									Do: types.DO_SET_NULL,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "dbid"."table1" ("id" INT8, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "dbid"."users"("username") ON UPDATE CASCADE ON DELETE SET DEFAULT, FOREIGN KEY ("id", "name") REFERENCES "dbid"."table2"("id", "username") ON UPDATE SET NULL ON DELETE SET NULL, PRIMARY KEY ("id"));`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generate.GenerateDDL("dbid", tt.args.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDDL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("GenerateDDL(): Got and Want have different lengths")
			}

			for i, statement := range got {
				want := tt.want[i]
				if !compareIgnoringWhitespace(statement, want) {
					t.Errorf("GenerateDDL() got = %v, want %v", got, tt.want)
				}

				err = postgres.CheckSyntaxReplaceDollar(statement)
				assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
			}
		})
	}
}

// there used to be a bug where the DDL generator would edit a table's primary key index,
// if one existed.  It would add an extra '\"' to the beginning and end of each column name.
func Test_PrimaryIndexModification(t *testing.T) {
	testTable := &types.Table{
		Name: "test",
		Columns: []*types.Column{
			{
				Name: "id1",
				Type: types.IntType,
			},
			{
				Name: "id2", // doing this to check composite primary keys
				Type: types.IntType,
			},
		},
		Indexes: []*types.Index{
			{
				Name: "primary",
				Columns: []string{
					"id1",
					"id2",
				},
				Type: types.PRIMARY,
			},
		},
	}

	_, err := generate.GenerateDDL("dbid", testTable)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check that the primary key index was not modified
	if testTable.Indexes[0].Columns[0] != "id1" {
		t.Errorf("primary key index was modified. Expected 'id1', got '%s'", testTable.Indexes[0].Columns[0])
	}

	if testTable.Indexes[0].Columns[1] != "id2" {
		t.Errorf("primary key index was modified. Expected 'id2', got '%s'", testTable.Indexes[0].Columns[1])
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

// compareIgnoringWhitespace compares two strings while ignoring whitespace characters.
func compareIgnoringWhitespace(a, b string) bool {
	aWithoutWhitespace := removeWhitespace(a)
	bWithoutWhitespace := removeWhitespace(b)

	return aWithoutWhitespace == bWithoutWhitespace
}
