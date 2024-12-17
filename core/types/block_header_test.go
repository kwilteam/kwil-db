package types

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBlockHeader_EncodeDecode(t *testing.T) {
	t.Run("encode and decode header", func(t *testing.T) {
		timestamp := time.Now().UTC().Truncate(time.Millisecond)
		original := &BlockHeader{
			Version:          1,
			Height:           100,
			NumTxns:          5,
			PrevHash:         Hash{1, 2, 3},
			PrevAppHash:      Hash{4, 5, 6},
			ValidatorSetHash: Hash{7, 8, 9},
			Timestamp:        timestamp,
			MerkleRoot:       Hash{10, 11, 12},
		}

		encoded := EncodeBlockHeader(original)
		decoded, err := DecodeBlockHeader(bytes.NewReader(encoded))
		require.NoError(t, err)

		require.Equal(t, original.Version, decoded.Version)
		require.Equal(t, original.Height, decoded.Height)
		require.Equal(t, original.NumTxns, decoded.NumTxns)
		require.Equal(t, original.PrevHash, decoded.PrevHash)
		require.Equal(t, original.PrevAppHash, decoded.PrevAppHash)
		require.Equal(t, original.ValidatorSetHash, decoded.ValidatorSetHash)
		require.Equal(t, original.Timestamp.UTC(), decoded.Timestamp)
		require.Equal(t, original.MerkleRoot, decoded.MerkleRoot)

		if !reflect.DeepEqual(original, decoded) {
			t.Errorf("expected %+v, got %+v", original, decoded)
		}
	})

	t.Run("decode with insufficient data", func(t *testing.T) {
		_, err := DecodeBlockHeader(bytes.NewReader([]byte{1, 2, 3}))
		require.Error(t, err)
	})

	t.Run("decode with empty reader", func(t *testing.T) {
		_, err := DecodeBlockHeader(bytes.NewReader(nil))
		require.Error(t, err)
	})

	t.Run("string representation", func(t *testing.T) {
		timestamp := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		header := &BlockHeader{
			Version:          1,
			Height:           100,
			NumTxns:          5,
			PrevHash:         Hash{1, 2, 3},
			PrevAppHash:      Hash{4, 5, 6},
			ValidatorSetHash: Hash{7, 8, 9},
			Timestamp:        timestamp,
			MerkleRoot:       Hash{10, 11, 12},
		}
		str := header.String()
		require.Contains(t, str, "Height: 100")
		require.Contains(t, str, "2023-01-01T00:00:00Z")
	})

	t.Run("write to failed writer", func(t *testing.T) {
		header := &BlockHeader{
			Version:    1,
			Height:     100,
			NumTxns:    5,
			Timestamp:  time.Now(),
			MerkleRoot: Hash{1, 2, 3},
		}
		err := header.writeBlockHeader(&failingWriter{})
		require.Error(t, err)
	})
}

type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}
