package order_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer/order"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

func Test_Ordering(t *testing.T) {
	type testCase struct {
		name                  string
		tables                []*types.Table
		selectCores           []*tree.SelectCore
		originalOrderingTerms []*tree.OrderingTerm
		expectedOrderingTerms []*tree.OrderingTerm
	}

	testCases := []testCase{
		{
			name:   "select star with no ordering",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("*").
					From("users").
					Build(),
			},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Table:  "users",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
		{
			name:   "simple ordering",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("id", "name").
					From("users").
					Build(),
			},
			originalOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Column: "name",
					},
					OrderType:    tree.OrderTypeDesc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Column: "name",
					},
					OrderType:    tree.OrderTypeDesc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "users",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
		{
			name:   "ordering with join and table alias",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("users.name", "posts.title").
					From("users").
					Join("posts", "p.user_id", "users.id", "p").
					Build(),
			},
			originalOrderingTerms: []*tree.OrderingTerm{},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Table:  "p",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "users",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
		{
			name:   "ordering with two tables with compound primary keys",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("f1.user_id", "f2.user_id").
					From("followers", "f1").
					Join("followers", "f2.follower_id", "f1.user_id", "f2").
					Build(),
			},
			originalOrderingTerms: []*tree.OrderingTerm{},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Table:  "f1",
						Column: "follower_id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "f1",
						Column: "user_id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "f2",
						Column: "follower_id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "f2",
						Column: "user_id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},

		{
			name:   "compound select, with column alias",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("users.name", "users.id").
					ColumnAs("uname", "uid").
					From("users").
					Compound(tree.CompoundOperatorTypeUnion).
					Build(),
				Select().
					Columns("users2.name", "users2.id").
					From("users", "users2").
					Build(),
			},
			originalOrderingTerms: []*tree.OrderingTerm{},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Column: "uname",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Column: "uid",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
		{
			name:   "compound select, with result star AND result column",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				Select().
					Columns("users.*", "users.id").
					From("users").
					Compound(tree.CompoundOperatorTypeUnion).
					Build(),
				Select().
					Columns("users2.*", "users2.id").
					From("users", "users2").
					Build(),
			},
			originalOrderingTerms: []*tree.OrderingTerm{},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Column: "name",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
				{
					Expression: &tree.ExpressionColumn{
						Table:  "users",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
		// we will test scoping by ensuring that a subquery does not include ordering terms for the parent query
		{
			name:   "verifying scoping",
			tables: defaultTables,
			selectCores: []*tree.SelectCore{
				{
					Columns: []tree.ResultColumn{
						&tree.ResultColumnStar{},
					},
					From: &tree.FromClause{
						JoinClause: &tree.JoinClause{
							TableOrSubquery: &tree.TableOrSubqueryTable{
								Name: "users",
							},
						},
					},
					Where: &tree.ExpressionSelect{
						Select: &tree.SelectStmt{
							SelectCores: []*tree.SelectCore{
								Select().
									Columns("posts.id").
									From("posts").
									Build(),
							},
						},
					},
				},
			},
			originalOrderingTerms: []*tree.OrderingTerm{},
			expectedOrderingTerms: []*tree.OrderingTerm{
				{
					Expression: &tree.ExpressionColumn{
						Table:  "users",
						Column: "id",
					},
					OrderType:    tree.OrderTypeAsc,
					NullOrdering: tree.NullOrderingTypeLast,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			walker := order.NewOrderWalker(tc.tables)

			// we test for nil orderBy, since a previous bug was caused by having an empty orderBy
			var orderBy *tree.OrderBy
			if tc.originalOrderingTerms != nil {
				orderBy = &tree.OrderBy{
					OrderingTerms: tc.originalOrderingTerms,
				}
			}

			selectStmt := &tree.SelectStmt{
				SelectCores: tc.selectCores,
				OrderBy:     orderBy,
			}

			err := selectStmt.Walk(walker)
			if err != nil {
				t.Fatal(err)
			}

			if tc.expectedOrderingTerms != nil {
				assert.EqualValues(t, tc.expectedOrderingTerms, selectStmt.OrderBy.OrderingTerms)
			}
		})
	}
}

var defaultTables = []*types.Table{
	{
		Name: "users",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "name",
				Type: types.TEXT,
			},
		},
		Indexes:     []*types.Index{},
		ForeignKeys: []*types.ForeignKey{},
	},
	{
		Name: "posts",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "user_id",
				Type: types.INT,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "title",
				Type: types.TEXT,
			},
		},
	},
	{
		Name: "followers",
		Columns: []*types.Column{
			{
				Name: "user_id",
				Type: types.INT,
			},
			{
				Name: "follower_id",
				Type: types.INT,
			},
		},
		Indexes: []*types.Index{
			{
				Name: "primary_key",
				Columns: []string{
					"user_id",
					"follower_id",
				},
				Type: types.PRIMARY,
			},
		},
	},
}
