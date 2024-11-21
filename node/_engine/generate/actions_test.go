package generate

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/assert"
)

func TestGenerateActionBody(t *testing.T) {
	tests := []struct {
		name            string
		action          *types.Action
		schema          *types.Schema
		pgSchema        string
		wantErr         bool
		wantGenActStmts []any
	}{
		{
			name: "invalid schema",
			action: &types.Action{
				Name: "test_action",
				Body: "SELECT * FROM nonexistent_table;",
			},
			schema:   nil,
			pgSchema: "test_schema",
			wantErr:  true,
		},
		{
			name: "invalid SQL syntax",
			action: &types.Action{
				Name: "invalid_syntax",
				Body: "SELECT * FROM;",
			},
			schema:   &types.Schema{},
			pgSchema: "test_schema",
			wantErr:  true,
		},
		{
			name: "valid select action",
			action: &types.Action{
				Name: "valid_action",
				Body: "SELECT id FROM users;",
			},
			schema: &types.Schema{
				Tables: []*types.Table{
					{
						Name: "users",
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
						},
					},
				},
			},
			pgSchema: "test_schema",
			wantErr:  false,
		},
		{
			name: "faithful typecast with repeat variable number in (*sqlGenerator).VisitExpressionVariable",
			action: &types.Action{
				Name:       "insert_typecasts",
				Parameters: []string{"$id", "$intval"},
				Body: `INSERT INTO id_and_int (
    id,
    intval
) VALUES (
    uuid_generate_v5('31276fd4-105f-4ff7-9f64-644942c14b79'::uuid, format('%s-%s', $id::text, $intval::text)),
    $intval::int
);`,
			},
			schema: &types.Schema{
				Tables: []*types.Table{
					{
						Name: "id_and_int",
						Columns: []*types.Column{
							{
								Name: "id",
								Type: types.UUIDType,
								Attributes: []*types.Attribute{
									{
										Type: types.PRIMARY_KEY,
									},
								},
							},
							{
								Name: "intval",
								Type: types.IntType,
							},
						},
					},
				},
			},
			pgSchema: "test_schema",
			wantErr:  false,
			wantGenActStmts: []any{
				&ActionSQL{
					Statement: `
INSERT INTO test_schema.id_and_int (id, intval) 
VALUES 
(uuid_generate_v5('31276fd4-105f-4ff7-9f64-644942c14b79'::UUID, format('%s-%s'::TEXT, $1::TEXT, $2::TEXT)), $2::INT8);`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmts, err := GenerateActionBody(tt.action, tt.schema, tt.pgSchema)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, stmts)

			if len(tt.wantGenActStmts) == 0 {
				return
			}
			for i, stmt := range stmts {
				switch st := stmt.(type) {
				case *ActionSQL:
					wantStmt, ok := tt.wantGenActStmts[i].(*ActionSQL)
					if !ok {
						t.Errorf("wanted *ActionSQL, got %T", tt.wantGenActStmts[i])
						return
					}
					assert.Equal(t, wantStmt.Statement, st.Statement)
				default:
				}
			}
		})
	}
}
