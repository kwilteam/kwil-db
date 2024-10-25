package types

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

func TestGetRawBlockTx(t *testing.T) {
	makeRawBlock := func(txns [][]byte) []byte {
		blk := NewBlock(1, Hash{1, 2, 3}, Hash{6, 7, 8}, time.Unix(1729890593, 0), txns)
		return EncodeBlock(blk)
	}

	t.Run("valid transaction index", func(t *testing.T) {
		txns := [][]byte{
			[]byte("tx1"),
			[]byte("transaction2"),
			[]byte("tx3"),
		}
		rawBlock := makeRawBlock(txns)

		tx, err := GetRawBlockTx(rawBlock, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(tx, txns[1]) {
			t.Errorf("got tx %x, want %x", tx, txns[1])
		}
	})

	t.Run("invalid transaction length", func(t *testing.T) {
		buf := new(bytes.Buffer)
		header := &BlockHeader{Height: 1, NumTxns: 1}
		buf.Write(EncodeBlockHeader(header))
		binary.Write(buf, binary.LittleEndian, uint32(1<<30)) // Very large tx length

		_, err := GetRawBlockTx(buf.Bytes(), 0)
		if err == nil {
			t.Error("expected error for invalid transaction length")
		}
	})

	t.Run("index out of range", func(t *testing.T) {
		txns := [][]byte{[]byte("tx1")}
		rawBlock := makeRawBlock(txns)

		_, err := GetRawBlockTx(rawBlock, 1)
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})

	t.Run("corrupted block data", func(t *testing.T) {
		txns := [][]byte{[]byte("tx1")}
		rawBlock := makeRawBlock(txns)
		corrupted := rawBlock[:len(rawBlock)-1]

		_, err := GetRawBlockTx(corrupted, 0)
		if err == nil {
			t.Error("expected error for corrupted block data")
		}
	})

	t.Run("empty block", func(t *testing.T) {
		rawBlock := makeRawBlock([][]byte{})

		_, err := GetRawBlockTx(rawBlock, 0)
		if err != ErrNotFound {
			t.Errorf("got err %v, want ErrNotFound", err)
		}
	})
}
