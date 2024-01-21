package attributes_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/testdata"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/attributes"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

func TestGetSelectCoreRelationAttributes(t *testing.T) {
	tests := []struct {
		name            string
		tables          []*common.Table
		stmt            string
		want            []*attributes.RelationAttribute
		resultTableCols []*common.Column
		wantErr         bool
	}{
		{
			name: "simple select",
			tables: []*common.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(common.INT, "users", "id"),
			},
			resultTableCols: []*common.Column{
				col("id", common.INT),
			},
		},
		{
			name: "simple select with alias",
			tables: []*common.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id AS user_id FROM users",
			want: []*attributes.RelationAttribute{
				tblColAlias(common.INT, "users", "id", "user_id"),
			},
			resultTableCols: []*common.Column{
				col("user_id", common.INT),
			},
		},
		{
			name: "test subquery is ignored",
			tables: []*common.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT id FROM users WHERE id IN (SELECT author_id FROM posts)",
			want: []*attributes.RelationAttribute{
				tblCol(common.INT, "users", "id"),
			},
			resultTableCols: []*common.Column{
				col("id", common.INT),
			},
		},
		{
			name: "test star, table star works",
			tables: []*common.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT users.*, * FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(common.INT, "users", "id"), // we expect them twice since it is defined twice
				tblCol(common.TEXT, "users", "username"),
				tblCol(common.INT, "users", "age"),
				tblCol(common.TEXT, "users", "address"),
				tblCol(common.INT, "users", "id"),
				tblCol(common.TEXT, "users", "username"),
				tblCol(common.INT, "users", "age"),
				tblCol(common.TEXT, "users", "address"),
			},
			resultTableCols: []*common.Column{
				col("id", common.INT),
				col("username", common.TEXT),
				col("age", common.INT),
				col("address", common.TEXT),
				col("id:1", common.INT),
				col("username:1", common.TEXT),
				col("age:1", common.INT),
				col("address:1", common.TEXT),
			},
		},
		{
			name: "test star, table star, literal, untabled column, and tabled column with alias work and join",
			tables: []*common.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT users.*, *, age, users.age AS the_age, 5 as the_literal_5 FROM users INNER JOIN posts ON users.id = posts.author_id",
			want: []*attributes.RelationAttribute{
				// all user columns from users.*
				tblCol(common.INT, "users", "id"),
				tblCol(common.TEXT, "users", "username"),
				tblCol(common.INT, "users", "age"),
				tblCol(common.TEXT, "users", "address"),

				// all user columns from *
				tblCol(common.INT, "users", "id"),
				tblCol(common.TEXT, "users", "username"),
				tblCol(common.INT, "users", "age"),
				tblCol(common.TEXT, "users", "address"),

				// all post columns from *
				tblCol(common.INT, "posts", "id"),
				tblCol(common.TEXT, "posts", "title"),
				tblCol(common.TEXT, "posts", "content"),
				tblCol(common.INT, "posts", "author_id"),
				tblCol(common.TEXT, "posts", "post_date"),

				// age
				tblCol(common.INT, "users", "age"),

				// users.age AS the_age
				tblColAlias(common.INT, "users", "age", "the_age"),

				// 5
				literal(common.INT, "5", "the_literal_5"),
			},
			resultTableCols: []*common.Column{
				col("id", common.INT),
				col("username", common.TEXT),
				col("age", common.INT),
				col("address", common.TEXT),
				col("id:1", common.INT),
				col("username:1", common.TEXT),
				col("age:1", common.INT),
				col("address:1", common.TEXT),
				col("id:2", common.INT),
				col("title", common.TEXT),
				col("content", common.TEXT),
				col("author_id", common.INT),
				col("post_date", common.TEXT),
				col("age:2", common.INT),
				col("the_age", common.INT),
				col("the_literal_5", common.INT),
			},
		},
		{
			name: "join with aliases",
			tables: []*common.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT u.id AS user_id, u.username AS username, count(p.id) AS post_count FROM users AS u LEFT JOIN posts AS p ON u.id = p.author_id GROUP BY u.id",
			want: []*attributes.RelationAttribute{
				tblColAlias(common.INT, "u", "id", "user_id"),
				tblColAlias(common.TEXT, "u", "username", "username"),
				{
					ResultExpression: &tree.ResultColumnExpression{
						Expression: &tree.ExpressionFunction{
							Function: &tree.FunctionCOUNT,
							Inputs: []tree.Expression{
								&tree.ExpressionColumn{
									Table:  "p",
									Column: "id",
								},
							},
						},
						Alias: "post_count",
					},
					Type: common.INT,
				},
			},
			resultTableCols: []*common.Column{
				col("user_id", common.INT),
				col("username", common.TEXT),
				col("post_count", common.INT),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := sqlparser.Parse(tt.stmt)
			if err != nil {
				t.Errorf("GetSelectCoreRelationAttributes() error = %v", err)
				return
			}
			selectStmt, okj := ast.(*tree.Select)
			if !okj {
				t.Errorf("test case %s is not a select statement", tt.name)
				return
			}

			got, err := attributes.GetSelectCoreRelationAttributes(selectStmt.SelectStmt.SelectCores[0], tt.tables)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSelectCoreRelationAttributes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Invalid length.  GetSelectCoreRelationAttributes() got = %v, want %v", got, tt.want)
				return
			}

			assert.ElementsMatch(t, got, tt.want, "GetSelectCoreRelationAttributes() got = %v, want %v", got, tt.want)

			genTable, err := attributes.TableFromAttributes("result_table", got, true)
			if err != nil {
				t.Errorf("GetSelectCoreRelationAttributes() error = %v", err)
				return
			}
			// check that the auto primary key works
			assert.Equal(t, len(tt.want), len(genTable.Indexes[0].Columns))

			assert.ElementsMatch(t, tt.resultTableCols, genTable.Columns, "GetSelectCoreRelationAttributes() got = %v, want %v", got, tt.want)

			sql, err := tree.SafeToSQL(selectStmt)
			assert.NoErrorf(t, err, "error converting query to SQL: %s", err)

			err = postgres.CheckSyntaxReplaceDollar(sql)
			assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
		})
	}
}

func tblCol(dataType common.DataType, tbl, column string) *attributes.RelationAttribute {
	return &attributes.RelationAttribute{
		ResultExpression: &tree.ResultColumnExpression{
			Expression: &tree.ExpressionColumn{
				Table:  tbl,
				Column: column,
			},
		},
		Type: dataType,
	}
}

func tblColAlias(dataType common.DataType, tbl, column, alias string) *attributes.RelationAttribute {
	return &attributes.RelationAttribute{
		ResultExpression: &tree.ResultColumnExpression{
			Expression: &tree.ExpressionColumn{
				Table:  tbl,
				Column: column,
			},
			Alias: alias,
		},
		Type: dataType,
	}
}

func literal(dataType common.DataType, lit string, alias string) *attributes.RelationAttribute {
	return &attributes.RelationAttribute{
		ResultExpression: &tree.ResultColumnExpression{
			Expression: &tree.ExpressionLiteral{
				Value: lit,
			},
			Alias: alias,
		},
		Type: dataType,
	}
}

func col(name string, datatype common.DataType) *common.Column {
	return &common.Column{
		Name: name,
		Type: datatype,
	}
}
