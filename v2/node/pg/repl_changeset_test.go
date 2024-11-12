package pg

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kwil/types"
)

func TestChangesetEntry_Serialize(t *testing.T) {
	tests := []struct {
		name string
		ce   *ChangesetEntry
	}{
		{
			name: "valid changeset entry",
			ce: &ChangesetEntry{
				RelationIdx: 1,
				OldTuple: []*TupleColumn{
					{
						ValueType: SerializedValue,
						Data:      []byte{2, 3, 4, 5},
					},
				},
				NewTuple: []*TupleColumn{
					{
						ValueType: SerializedValue,
						Data:      []byte{4, 5, 6, 7},
					},
				},
			},
		},
		{
			name: "changeset entry with empty old",
			ce: &ChangesetEntry{
				RelationIdx: 1,
				OldTuple:    []*TupleColumn{}, // nil does not round trip in RLP!
				NewTuple: []*TupleColumn{
					{
						ValueType: SerializedValue,
						Data:      []byte{4, 5, 6, 7},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test round trip with MarshalBinary and UnmarshalBinary
			bts, err := tt.ce.MarshalBinary()
			require.NoError(t, err)

			// Deserialize and compare
			newCE := &ChangesetEntry{}
			err = newCE.UnmarshalBinary(bts)
			require.NoError(t, err)
			assert.Equal(t, tt.ce, newCE)

			// Now as a prefixed element in a stream
			var buf bytes.Buffer
			err = StreamElement(&buf, tt.ce)
			require.NoError(t, err)

			csStream := buf.Bytes()
			csType, csSize := DecodeStreamPrefix([5]byte(csStream[:5]))
			assert.Equal(t, ChangesetEntryType, csType)
			assert.Equal(t, int(csSize), len(bts))
		})
	}
}

func TestRelation_Serialize(t *testing.T) {
	tests := []struct {
		name string
		r    *Relation
	}{
		{
			name: "valid relation",
			r: &Relation{
				Schema: "ns",
				Table:  "table",
				Columns: []*Column{
					{Name: "a", Type: types.IntType},
					{Name: "b", Type: types.TextType},
				},
			},
		},
		{
			name: "changeset entry with no schema",
			r: &Relation{
				Table: "tablex",
				Columns: []*Column{
					{Name: "a", Type: types.BlobType},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First test round trip with MarshalBinary and UnmarshalBinary
			bts, err := tt.r.MarshalBinary()
			require.NoError(t, err)

			// Deserialize and compare
			rel := &Relation{}
			err = rel.UnmarshalBinary(bts)
			require.NoError(t, err)
			assert.Equal(t, tt.r, rel)

			// Now as a prefixed element in a stream
			var buf bytes.Buffer
			err = StreamElement(&buf, tt.r)
			require.NoError(t, err)

			csStream := buf.Bytes()
			csType, csSize := DecodeStreamPrefix([5]byte(csStream[:5]))
			assert.Equal(t, RelationType, csType)
			assert.Equal(t, int(csSize), len(bts))
		})
	}
}
