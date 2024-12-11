package types

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/types"

	"github.com/stretchr/testify/require"
)

func TestGetRawBlockTx(t *testing.T) {
	privKey, pubKey, err := crypto.GenerateSecp256k1Key(nil)
	require.NoError(t, err)

	makeRawBlock := func(txns [][]byte) []byte {
		blk := types.NewBlock(1, Hash{1, 2, 3}, Hash{6, 7, 8}, Hash{}, time.Unix(1729890593, 0), txns)
		err := blk.Sign(privKey)
		require.NoError(t, err)
		return types.EncodeBlock(blk)
	}

	t.Run("valid block signature", func(t *testing.T) {
		txns := [][]byte{[]byte("tx1")}
		rawBlock := makeRawBlock(txns)
		blk, err := types.DecodeBlock(rawBlock)
		require.NoError(t, err)

		valid, err := blk.VerifySignature(pubKey)
		require.NoError(t, err)
		require.True(t, valid)
	})

	t.Run("valid transaction index", func(t *testing.T) {
		txns := [][]byte{
			[]byte("tx1"),
			[]byte("transaction2"),
			[]byte("tx3"),
		}
		rawBlock := makeRawBlock(txns)
		tx, err := types.GetRawBlockTx(rawBlock, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(tx, txns[1]) {
			t.Errorf("got tx %x, want %x", tx, txns[1])
		}
	})

	t.Run("invalid transaction length", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := &types.BlockHeader{Height: 1, NumTxns: 1}
		buf.Write(types.EncodeBlockHeader(header))
		binary.Write(buf, binary.LittleEndian, uint32(1<<30)) // Very large tx length

		_, err := types.GetRawBlockTx(buf.Bytes(), 0)
		if err == nil {
			t.Error("expected error for invalid transaction length")
		}
	})

	t.Run("index out of range", func(t *testing.T) {
		txns := [][]byte{[]byte("tx1")}
		rawBlock := makeRawBlock(txns)

		_, err := types.GetRawBlockTx(rawBlock, 1)
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("corrupted block data", func(t *testing.T) {
		txns := [][]byte{[]byte("tx1")}
		rawBlock := makeRawBlock(txns)
		blk, err := types.DecodeBlock(rawBlock)
		require.NoError(t, err)

		sigLen := len(blk.Signature) + 4
		corrupted := rawBlock[:len(rawBlock)-1-sigLen]

		_, err = types.GetRawBlockTx(corrupted, 0)
		if err == nil {
			t.Error("expected error for corrupted block data")
		}
	})

	t.Run("empty block", func(t *testing.T) {
		rawBlock := makeRawBlock([][]byte{})

		_, err := types.GetRawBlockTx(rawBlock, 0)
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})
}
func TestCalcMerkleRoot(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		root := types.CalcMerkleRoot([]Hash{})
		if root != (Hash{}) {
			t.Errorf("empty slice should return zero hash, got %x", root)
		}
	})

	t.Run("single leaf", func(t *testing.T) {
		leaf := Hash{1, 2, 3, 4}
		root := types.CalcMerkleRoot([]Hash{leaf})
		if root != leaf {
			t.Errorf("single leaf should return same hash, got %x, want %x", root, leaf)
		}
	})

	t.Run("two leaves", func(t *testing.T) {
		leaf1 := Hash{1, 2, 3, 4}
		leaf2 := Hash{5, 6, 7, 8}
		root := types.CalcMerkleRoot([]Hash{leaf1, leaf2})

		var buf [HashLen * 2]byte
		copy(buf[:HashLen], leaf1[:])
		copy(buf[HashLen:], leaf2[:])
		expected := types.HashBytes(buf[:])

		if root != expected {
			t.Errorf("got root %x, want %x", root, expected)
		}
	})

	t.Run("three leaves", func(t *testing.T) {
		leaves := []Hash{
			{1, 1, 1, 1},
			{2, 2, 2, 2},
			{3, 3, 3, 3},
		}
		types.CalcMerkleRoot(leaves)

		// Verify original slice not modified
		if leaves[2] != (Hash{3, 3, 3, 3}) {
			t.Error("original slice was modified")
		}
	})

	t.Run("four leaves", func(t *testing.T) {
		leaves := []Hash{
			{1, 1, 1, 1},
			{2, 2, 2, 2},
			{3, 3, 3, 3},
			{4, 4, 4, 4},
		}
		root1 := types.CalcMerkleRoot(leaves)

		// Calculate same root with modified order
		leaves[0], leaves[1] = leaves[1], leaves[0]
		root2 := types.CalcMerkleRoot(leaves)

		if root1 == root2 {
			t.Error("roots should differ when leaf order changes")
		}
	})

	t.Run("preserve input", func(t *testing.T) {
		original := []Hash{
			{1, 1, 1, 1},
			{2, 2, 2, 2},
		}
		originalCopy := make([]Hash, len(original))
		copy(originalCopy, original)

		types.CalcMerkleRoot(original)

		for i := range original {
			if original[i] != originalCopy[i] {
				t.Errorf("input slice was modified at index %d", i)
			}
		}
	})
}
