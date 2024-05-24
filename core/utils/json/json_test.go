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
