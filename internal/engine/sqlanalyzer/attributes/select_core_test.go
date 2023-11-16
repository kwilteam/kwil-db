package attributes_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/attributes"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/engine/types/testdata"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
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
				tblCol(types.INT, "users", "id"),
			},
			resultTableCols: []*types.Column{
				col("id", types.INT),
			},
		},
		{
			name: "simple select with alias",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id AS user_id FROM users",
			want: []*attributes.RelationAttribute{
				tblColAlias(types.INT, "users", "id", "user_id"),
			},
			resultTableCols: []*types.Column{
				col("user_id", types.INT),
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
				tblCol(types.INT, "users", "id"),
			},
			resultTableCols: []*types.Column{
				col("id", types.INT),
			},
		},
		{
			name: "test star, table star works",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT users.*, * FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(types.INT, "users", "id"), // we expect them twice since it is defined twice
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.BLOB, "users", "address"),
				tblCol(types.INT, "users", "id"),
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.BLOB, "users", "address"),
			},
			resultTableCols: []*types.Column{
				col("id", types.INT),
				col("username", types.TEXT),
				col("age", types.INT),
				col("address", types.BLOB),
				col("id:1", types.INT),
				col("username:1", types.TEXT),
				col("age:1", types.INT),
				col("address:1", types.BLOB),
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
				tblCol(types.INT, "users", "id"),
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.BLOB, "users", "address"),

				// all user columns from *
				tblCol(types.INT, "users", "id"),
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.BLOB, "users", "address"),

				// all post columns from *
				tblCol(types.INT, "posts", "id"),
				tblCol(types.TEXT, "posts", "title"),
				tblCol(types.TEXT, "posts", "content"),
				tblCol(types.INT, "posts", "author_id"),
				tblCol(types.TEXT, "posts", "post_date"),

				// age
				tblCol(types.INT, "users", "age"),

				// users.age AS the_age
				tblColAlias(types.INT, "users", "age", "the_age"),

				// 5
				literal(types.INT, "5", "the_literal_5"),
			},
			resultTableCols: []*types.Column{
				col("id", types.INT),
				col("username", types.TEXT),
				col("age", types.INT),
				col("address", types.BLOB),
				col("id:1", types.INT),
				col("username:1", types.TEXT),
				col("age:1", types.INT),
				col("address:1", types.BLOB),
				col("id:2", types.INT),
				col("title", types.TEXT),
				col("content", types.TEXT),
				col("author_id", types.INT),
				col("post_date", types.TEXT),
				col("age:2", types.INT),
				col("the_age", types.INT),
				col("the_literal_5", types.INT),
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
				tblColAlias(types.INT, "u", "id", "user_id"),
				tblColAlias(types.TEXT, "u", "username", "username"),
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
					Type: types.INT,
				},
			},
			resultTableCols: []*types.Column{
				col("user_id", types.INT),
				col("username", types.TEXT),
				col("post_count", types.INT),
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
		})
	}
}

func tblCol(dataType types.DataType, tbl, column string) *attributes.RelationAttribute {
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

func tblColAlias(dataType types.DataType, tbl, column, alias string) *attributes.RelationAttribute {
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

func literal(dataType types.DataType, lit string, alias string) *attributes.RelationAttribute {
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

func col(name string, datatype types.DataType) *types.Column {
	return &types.Column{
		Name: name,
		Type: datatype,
	}
}
