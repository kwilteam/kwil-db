package attributes_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/common/testdata"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/attributes"
	sqlparser "github.com/kwilteam/kwil-db/internal/parse/sql"
	"github.com/kwilteam/kwil-db/internal/parse/sql/postgres"
	"github.com/kwilteam/kwil-db/internal/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

func TestGetSelectCoreRelationAttributes(t *testing.T) {
	tests := []struct {
		name            string
		tables          []*types.Table
		stmt            string
		want            []*attributes.RelationAttribute
		resultTableCols []*types.Column
		wantErr         bool
	}{
		{
			name: "simple select",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(types.IntType, "users", "id"),
			},
			resultTableCols: []*types.Column{
				col("id", types.IntType),
			},
		},
		{
			name: "simple select with alias",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id AS user_id FROM users",
			want: []*attributes.RelationAttribute{
				tblColAlias(types.IntType, "users", "id", "user_id"),
			},
			resultTableCols: []*types.Column{
				col("user_id", types.IntType),
			},
		},
		{
			name: "test subquery is ignored",
			tables: []*types.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT id FROM users WHERE id IN (SELECT author_id FROM posts)",
			want: []*attributes.RelationAttribute{
				tblCol(types.IntType, "users", "id"),
			},
			resultTableCols: []*types.Column{
				col("id", types.IntType),
			},
		},
		{
			name: "test star, table star works",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT users.*, * FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(types.IntType, "users", "id"), // we expect them twice since it is defined twice
				tblCol(types.TextType, "users", "username"),
				tblCol(types.IntType, "users", "age"),
				tblCol(types.TextType, "users", "address"),
				tblCol(types.IntType, "users", "id"),
				tblCol(types.TextType, "users", "username"),
				tblCol(types.IntType, "users", "age"),
				tblCol(types.TextType, "users", "address"),
			},
			resultTableCols: []*types.Column{
				col("id", types.IntType),
				col("username", types.TextType),
				col("age", types.IntType),
				col("address", types.TextType),
				col("id:1", types.IntType),
				col("username:1", types.TextType),
				col("age:1", types.IntType),
				col("address:1", types.TextType),
			},
		},
		{
			name: "test star, table star, literal, untabled column, and tabled column with alias work and join",
			tables: []*types.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT users.*, *, age, users.age AS the_age, 5 as the_literal_5 FROM users INNER JOIN posts ON users.id = posts.author_id",
			want: []*attributes.RelationAttribute{
				// all user columns from users.*
				tblCol(types.IntType, "users", "id"),
				tblCol(types.TextType, "users", "username"),
				tblCol(types.IntType, "users", "age"),
				tblCol(types.TextType, "users", "address"),

				// all user columns from *
				tblCol(types.IntType, "users", "id"),
				tblCol(types.TextType, "users", "username"),
				tblCol(types.IntType, "users", "age"),
				tblCol(types.TextType, "users", "address"),

				// all post columns from *
				tblCol(types.IntType, "posts", "id"),
				tblCol(types.TextType, "posts", "title"),
				tblCol(types.TextType, "posts", "content"),
				tblCol(types.IntType, "posts", "author_id"),
				tblCol(types.TextType, "posts", "post_date"),

				// age
				tblCol(types.IntType, "users", "age"),

				// users.age AS the_age
				tblColAlias(types.IntType, "users", "age", "the_age"),

				// 5
				{
					ResultExpression: &tree.ResultColumnExpression{
						Expression: &tree.ExpressionNumericLiteral{
							Value: 5,
						},
						Alias: "the_literal_5",
					},
					Type: types.IntType,
				},
			},
			resultTableCols: []*types.Column{
				col("id", types.IntType),
				col("username", types.TextType),
				col("age", types.IntType),
				col("address", types.TextType),
				col("id:1", types.IntType),
				col("username:1", types.TextType),
				col("age:1", types.IntType),
				col("address:1", types.TextType),
				col("id:2", types.IntType),
				col("title", types.TextType),
				col("content", types.TextType),
				col("author_id", types.IntType),
				col("post_date", types.TextType),
				col("age:2", types.IntType),
				col("the_age", types.IntType),
				col("the_literal_5", types.IntType),
			},
		},
		{
			name: "join with aliases",
			tables: []*types.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT u.id AS user_id, u.username AS username, count(p.id) AS post_count FROM users AS u LEFT JOIN posts AS p ON u.id = p.author_id GROUP BY u.id",
			want: []*attributes.RelationAttribute{
				tblColAlias(types.IntType, "u", "id", "user_id"),
				tblColAlias(types.TextType, "u", "username", "username"),
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
					Type: types.IntType,
				},
			},
			resultTableCols: []*types.Column{
				col("user_id", types.IntType),
				col("username", types.TextType),
				col("post_count", types.IntType),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmt, err := sqlparser.Parse(tt.stmt)
			if err != nil {
				t.Errorf("GetSelectCoreRelationAttributes() error = %v", err)
				return
			}
			selectStmt, okj := stmt.(*tree.SelectStmt)
			if !okj {
				t.Errorf("test case %s is not a select statement", tt.name)
				return
			}

			got, err := attributes.GetSelectCoreRelationAttributes(selectStmt.Stmt.SimpleSelects[0], tt.tables)
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

func tblCol(dataType *types.DataType, tbl, column string) *attributes.RelationAttribute {
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

func tblColAlias(dataType *types.DataType, tbl, column, alias string) *attributes.RelationAttribute {
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

func col(name string, datatype *types.DataType) *types.Column {
	return &types.Column{
		Name: name,
		Type: datatype,
	}
}
