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
)

func TestGetSelectCoreRelationAttributes(t *testing.T) {
	tests := []struct {
		name           string
		tables         []*types.Table
		stmt           string
		want           []*tree.ResultColumnExpression
		wantInequality bool
		wantErr        bool
	}{
		{
			name: "simple select",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id FROM users",
			want: []*tree.ResultColumnExpression{
				tblCol("users", "id"),
			},
		},
		{
			name: "simple select - failure",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id FROM users",
			want: []*tree.ResultColumnExpression{
				tblCol("users", "name"),
			},
			wantInequality: true,
		},
		{
			name: "simple select with alias",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT id AS user_id FROM users",
			want: []*tree.ResultColumnExpression{
				tblColAlias("users", "id", "user_id"),
			},
		},
		{
			name: "test subquery is ignored",
			tables: []*types.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT id FROM users WHERE id IN (SELECT user_id FROM posts)",
			want: []*tree.ResultColumnExpression{
				tblCol("users", "id"),
			},
		},
		{
			name: "test star, table star works",
			tables: []*types.Table{
				testdata.TableUsers,
			},
			stmt: "SELECT users.*, * FROM users",
			want: []*tree.ResultColumnExpression{
				tblCol("users", "id"), // we expect them twice since it is defined twice
				tblCol("users", "username"),
				tblCol("users", "age"),
				tblCol("users", "address"),
				tblCol("users", "id"),
				tblCol("users", "username"),
				tblCol("users", "age"),
				tblCol("users", "address"),
			},
		},
		{
			name: "test star, table star, literal, untabled column, and tabled column with alias work and join",
			tables: []*types.Table{
				testdata.TableUsers,
				testdata.TablePosts,
			},
			stmt: "SELECT users.*, *, age, users.age AS the_age, 5 FROM users INNER JOIN posts ON users.id = posts.author_id",
			want: []*tree.ResultColumnExpression{
				// all user columns from users.*
				tblCol("users", "id"),
				tblCol("users", "username"),
				tblCol("users", "age"),
				tblCol("users", "address"),

				// all user columns from *
				tblCol("users", "id"),
				tblCol("users", "username"),
				tblCol("users", "age"),
				tblCol("users", "address"),

				// all post columns from *
				tblCol("posts", "id"),
				tblCol("posts", "title"),
				tblCol("posts", "content"),
				tblCol("posts", "author_id"),
				tblCol("posts", "post_date"),

				// age
				tblCol("users", "age"),

				// users.age AS the_age
				tblColAlias("users", "age", "the_age"),

				// 5
				literal("5"),
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
				if !cmp.Equal(*g, *tt.want[i], cmpopts.IgnoreUnexported(tree.ExpressionColumn{}, tree.ExpressionLiteral{}), cmpopts.EquateEmpty()) {
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
		})
	}
}

func tblCol(tbl, col string) *tree.ResultColumnExpression {
	return &tree.ResultColumnExpression{
		Expression: &tree.ExpressionColumn{
			Table:  tbl,
			Column: col,
		},
	}
}

func tblColAlias(tbl, col, alias string) *tree.ResultColumnExpression {
	return &tree.ResultColumnExpression{
		Expression: &tree.ExpressionColumn{
			Table:  tbl,
			Column: col,
		},
		Alias: alias,
	}
}

func literal(lit string) *tree.ResultColumnExpression {
	return &tree.ResultColumnExpression{
		Expression: &tree.ExpressionLiteral{
			Value: lit,
		},
	}
}
