package attributes_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlanalyzer/attributes"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser"
	"github.com/kwilteam/kwil-db/pkg/engine/sqlparser/tree"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
	"github.com/stretchr/testify/assert"
)

func TestGetSelectCoreRelationAttributes(t *testing.T) {
	tests := []struct {
		name            string
		tables          []*types.Table
		stmt            string
		want            []*attributes.RelationAttribute
		resultTableCols []*types.Column
		// wantInequality is true if we want the test to fail if the result is equal to want
		wantInequality bool
		wantErr        bool
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
			name: "simple select - failure",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id FROM users",
			want: []*attributes.RelationAttribute{
				tblCol(types.TEXT, "users", "name"),
			},
			wantInequality: true,
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
				tblCol(types.TEXT, "users", "address"),
				tblCol(types.INT, "users", "id"),
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.TEXT, "users", "address"),
			},
			resultTableCols: []*types.Column{
				col("id", types.INT),
				col("username", types.TEXT),
				col("age", types.INT),
				col("address", types.TEXT),
				col("id:1", types.INT),
				col("username:1", types.TEXT),
				col("age:1", types.INT),
				col("address:1", types.TEXT),
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
				tblCol(types.TEXT, "users", "address"),

				// all user columns from *
				tblCol(types.INT, "users", "id"),
				tblCol(types.TEXT, "users", "username"),
				tblCol(types.INT, "users", "age"),
				tblCol(types.TEXT, "users", "address"),

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
				col("address", types.TEXT),
				col("id:1", types.INT),
				col("username:1", types.TEXT),
				col("age:1", types.INT),
				col("address:1", types.TEXT),
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

			same := true
			incorrectIdx := -1
			for i, g := range got {
				if !cmp.Equal(*g, *tt.want[i], cmpopts.IgnoreUnexported(tree.ResultColumnExpression{}, tree.AnySQLFunction{}, tree.AggregateFunc{}, tree.ExpressionFunction{}, tree.ExpressionColumn{}, tree.ExpressionLiteral{}), cmpopts.EquateEmpty()) {
					same = false
					incorrectIdx = i
					break
				}
			}

			if same != !tt.wantInequality {
				t.Errorf("GetSelectCoreRelationAttributes() got = %v, want %v", got, tt.want)
				if incorrectIdx != -1 {
					t.Errorf("Incorrect index: %d", incorrectIdx)
				}
			}

			if tt.wantInequality {
				return
			}

			genTable, err := attributes.TableFromAttributes("result_table", got, true)
			if err != nil {
				t.Errorf("GetSelectCoreRelationAttributes() error = %v", err)
				return
			}
			// check that the auto primary key works
			assert.Equal(t, len(tt.want), len(genTable.Indexes[0].Columns))

			// check that the columns are correct
			if !cmp.Equal(tt.resultTableCols, genTable.Columns, cmpopts.IgnoreSliceElements(func(v int) bool { return true })) {
				t.Errorf("GetSelectCoreRelationAttributes() got = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func tblCol(dataType types.DataType, tbl, col string) *attributes.RelationAttribute {
	return &attributes.RelationAttribute{
		ResultExpression: &tree.ResultColumnExpression{
			Expression: &tree.ExpressionColumn{
				Table:  tbl,
				Column: col,
			},
		},
		Type: dataType,
	}
}

func tblColAlias(dataType types.DataType, tbl, col, alias string) *attributes.RelationAttribute {
	return &attributes.RelationAttribute{
		ResultExpression: &tree.ResultColumnExpression{
			Expression: &tree.ExpressionColumn{
				Table:  tbl,
				Column: col,
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
