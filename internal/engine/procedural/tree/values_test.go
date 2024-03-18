package tree_test

import (
	"testing"

	coreTypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/procedural/tree"
	types "github.com/kwilteam/kwil-db/internal/engine/procedural/types"
)

// TestValues tests that values are properly converted to strings for use in SQL statements.
func Test_Values(t *testing.T) {
	info := &tree.SystemInfo{
		Schemas: map[string]*tree.SchemaInfo{
			"public": {},
			"custom1": {
				Types: map[string]*types.CompositeTypeDefinition{
					"user": {
						Name: "user",
						Fields: []*types.CompositeTypeField{
							{
								Name: "id",
								Type: types.TypeUUID,
							},
							{
								Name: "name",
								Type: types.TypeText,
							},
						},
					},
					"app": {
						Name: "app",
						Fields: []*types.CompositeTypeField{
							{
								Name: "id",
								Type: types.TypeUUID,
							},
							{
								Name: "name",
								Type: types.TypeText,
							},
							{
								Name: "users",
								Type: &types.ArrayType{
									Type: &types.CustomType{
										Schema: "custom1",
										Name:   "user",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	type testcase struct {
		name  string
		value tree.Value
		want  string
	}

	tests := []testcase{
		{
			name: "Nested Composite Type",
			value: &tree.CompositeValue{
				Type: &types.CustomType{
					Schema: "custom1",
					Name:   "app",
				},
				Values: map[string]tree.Value{
					"id":   uuid("a"),
					"name": text("app1"),
					"users": &tree.ArrayValue{
						DataType: &types.CustomType{
							Schema: "custom1",
							Name:   "user",
						},
						Values: []tree.Value{
							&tree.CompositeValue{
								Type: &types.CustomType{
									Schema: "custom1",
									Name:   "user",
								},
								Values: map[string]tree.Value{
									"id":   uuid("b"),
									"name": text("user1"),
								},
							},
							&tree.CompositeValue{
								Type: &types.CustomType{
									Schema: "custom1",
									Name:   "user",
								},
								Values: map[string]tree.Value{
									"id":   uuid("c"),
									"name": text("user2"),
								},
							},
						},
					},
				},
			},
			want: `ROW('c430dbfc-28a9-508a-aee9-e6a2e1d81bb1', 'app1', ARRAY[ROW('8d89d4b1-aa51-554e-bc23-49a012d45ed0', 'user1')::custom1.user, ROW('d90dc498-8435-52b7-8994-e17d505dff32', 'user2')::custom1.user])::custom1.app`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.value.MarshalPG(info)
			if err != nil {
				t.Errorf("MarshalPG() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("MarshalPG() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func uuid(seed string) *tree.UUIDValue {
	a := tree.UUIDValue(coreTypes.NewUUIDV5([]byte(seed)))
	return &a
}

func text(s string) *tree.TextValue {
	a := tree.TextValue(s)
	return &a
}
