package json

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_convertJsonNumbers(t *testing.T) {

	tests := []struct {
		name string
		val  any
		want any
	}{
		{
			name: "number",
			val:  json.Number("123"),
			want: int64(123),
		},
		{
			name: "object",
			val: map[string]any{
				"key": json.Number("123"),
				"val": []map[string]any{
					{
						"key": json.Number("123"),
					},
					{
						"key": json.Number("456"),
						"val": []map[string]any{
							{
								"key": json.Number("789"),
							},
						},
					},
				},
			},
			want: map[string]any{
				"key": int64(123),
				"val": []map[string]any{
					{
						"key": int64(123),
					},
					{
						"key": int64(456),
						"val": []map[string]any{
							{
								"key": int64(789),
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := convertJsonNumbers(tt.val)
			require.EqualValues(t, tt.want, a)
		})
	}
}

// tests round-trip
type genericTest[T UnmarshalObject] struct {
	name string
	val  T
	want T
}

func (g *genericTest[T]) Name() string {
	return g.name
}

func (g *genericTest[T]) run(t *testing.T) {
	b, err := json.Marshal(g.val)
	require.NoError(t, err)

	res, err := UnmarshalMapWithoutFloat[T](b)
	require.NoError(t, err)

	require.EqualValues(t, g.want, res)
}

type testable interface {
	Name() string
	run(*testing.T)
}

func TestUnmarshalMapWithoutFloat(t *testing.T) {
	for _, tt := range []testable{
		&genericTest[map[string]any]{
			name: "map",
			val: map[string]any{
				"key": json.Number("123"),
			},
			want: map[string]any{
				"key": int64(123),
			},
		},
		&genericTest[[]map[string]any]{
			name: "array of maps",
			val: []map[string]any{
				{
					"key": []map[string]any{
						{
							"key": json.Number("123"),
						},
					},
				},
				{
					"key2": []any{int64(1), int64(2)},
				},
			},
			want: []map[string]any{
				{
					"key": []any{
						map[string]any{
							"key": int64(123),
						},
					},
				},
				{
					"key2": []any{int64(1), int64(2)},
				},
			},
		},
		&genericTest[[]any]{
			name: "array of any",
			val: []any{
				json.Number("123"),
				[]map[string]any{
					{
						"key": json.Number("123"),
					},
				},
			},
			want: []any{
				int64(123),
				[]any{
					map[string]any{
						"key": int64(123),
					},
				},
			},
		},
		&genericTest[string]{
			name: "string",
			val:  "123",
			want: "123",
		},
		&genericTest[int64]{
			name: "int64",
			val:  int64(123),
			want: int64(123),
		},
	} {
		t.Run(tt.Name(), tt.run)
	}
}

// regression test
func Test_EmptyBts(t *testing.T) {
	bts := []byte("{}")

	var aa map[string]any
	err := json.Unmarshal(bts, &aa)
	require.NoError(t, err)

	_, err = UnmarshalMapWithoutFloat[map[string]any](bts)
	require.NoError(t, err)
}
