package memstore

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types"
)

func fakeAppHash(height int64) types.Hash {
	return types.HashBytes(binary.LittleEndian.AppendUint64(nil, uint64(height)))
}

func newTx(nonce uint64, sender string) *ktypes.Transaction {
	return &ktypes.Transaction{
		Signature: &auth.Signature{},
		Body: &ktypes.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}
}

func createTestBlock(t *testing.T, height int64, numTxns int) (*ktypes.Block, types.Hash) {
	txs := make([]*ktypes.Transaction, numTxns)
	txns := make([][]byte, numTxns)
	for i := range numTxns {
		tx := newTx(uint64(i)+uint64(height), "sender")
		tx.Body.Payload = []byte(strings.Repeat("data", 1000))
		rawTx, err := tx.MarshalBinary()
		if err != nil {
			panic(err)
		}
		txs[i] = tx
		txns[i] = rawTx
	}
	blk := ktypes.NewBlock(height, types.Hash{2, 3, 4}, types.Hash{6, 7, 8}, types.Hash{5, 5, 5},
		time.Unix(1729723553+height, 0), txs)
	return blk, fakeAppHash(height)
}

func TestMemBS_StoreAndGet(t *testing.T) {
	bs := NewMemBS()

	block, appHash := createTestBlock(t, 1, 2)

	err := bs.Store(block, appHash)
	if err != nil {
		t.Fatal(err)
	}

	gotBlock, gotAppHash, err := bs.Get(block.Hash())
	if err != nil {
		t.Fatal(err)
	}

	if gotBlock.Header.Height != block.Header.Height {
		t.Errorf("got height %d, want %d", gotBlock.Header.Height, block.Header.Height)
	}

	if gotAppHash != appHash {
		t.Errorf("got app hash %v, want %v", gotAppHash, appHash)
	}
}

func TestMemBS_GetByHeight(t *testing.T) {
	bs := NewMemBS()

	blocks := []*ktypes.Block{
		{Header: &ktypes.BlockHeader{Height: 1}},
		{Header: &ktypes.BlockHeader{Height: 2}},
		{Header: &ktypes.BlockHeader{Height: 3}},
	}

	for i, block := range blocks {
		appHash := types.Hash{byte(i + 1)}
		if err := bs.Store(block, appHash); err != nil {
			t.Fatal(err)
		}
	}

	hash, block, appHash, err := bs.GetByHeight(2)
	if err != nil {
		t.Fatal(err)
	}

	if block.Header.Height != 2 {
		t.Errorf("got height %d, want 2", block.Header.Height)
	}

	if appHash != (types.Hash{2}) {
		t.Errorf("got app hash %v, want %v", appHash, types.Hash{2})
	}

	if hash != block.Hash() {
		t.Errorf("got hash %v, want %v", hash, block.Hash())
	}
}

func TestMemBS_Best(t *testing.T) {
	bs := NewMemBS()

	blocks := []*ktypes.Block{
		{Header: &ktypes.BlockHeader{Height: 1}},
		{Header: &ktypes.BlockHeader{Height: 3}},
		{Header: &ktypes.BlockHeader{Height: 2}},
	}

	for i, block := range blocks {
		appHash := types.Hash{byte(i + 1)}
		if err := bs.Store(block, appHash); err != nil {
			t.Fatal(err)
		}
	}

	height, hash, appHash := bs.Best()
	if height != 3 {
		t.Errorf("got height %d, want 3", height)
	}

	expectedBlock := blocks[1]
	if hash != expectedBlock.Hash() {
		t.Errorf("got hash %v, want %v", hash, expectedBlock.Hash())
	}

	if appHash != (types.Hash{2}) {
		t.Errorf("got app hash %v, want %v", appHash, types.Hash{2})
	}
}

func TestMemBS_StoreAndGetTx(t *testing.T) {
	bs := NewMemBS()

	// prevHash := types.Hash{7, 8, 9}
	// appHash := types.Hash{4, 2, 1}
	// valSetHash := types.Hash{4, 5, 6}

	// tx1 := []byte("tx1")
	// tx2 := []byte("tx2")
	// txns := [][]byte{tx1, tx2}
	// block := types.NewBlock(1, prevHash, appHash, valSetHash, time.Unix(123456789, 0), txns)
	block, _ := createTestBlock(t, 1, 2)
	tx1 := block.Txns[0]

	if err := bs.Store(block, types.Hash{1, 2, 3}); err != nil {
		t.Fatal(err)
	}

	txHash := tx1.Hash()
	gotTx, height, hash, idx, err := bs.GetTx(txHash)
	if err != nil {
		t.Fatal(err)
	}

	if height != 1 {
		t.Errorf("got height %d, want 1", height)
	}

	rawTx1, _ := tx1.MarshalBinary()
	gotRawTx, err := gotTx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotRawTx, rawTx1) {
		t.Errorf("got tx %x, want %x", gotRawTx, rawTx1)
	}

	if blkHash := block.Hash(); hash != blkHash {
		t.Errorf("incorrect block hash, got %v, expected %v", hash, blkHash)
	}

	if idx != 0 {
		t.Errorf("got index %d, want 0", idx)
	}
}

func TestMemBS_PreFetch(t *testing.T) {
	bs := NewMemBS()
	block := &ktypes.Block{Header: &ktypes.BlockHeader{Height: 1}}

	if err := bs.Store(block, types.Hash{}); err != nil {
		t.Fatal(err)
	}

	needFetch, done := bs.PreFetch(block.Hash())
	if needFetch {
		t.Fatal("expected no fetch needed for existing block")
	}
	done()

	newHash := types.Hash{1}
	needFetch, done = bs.PreFetch(newHash)
	if !needFetch {
		t.Error("expected fetch needed for new block")
	}

	if !bs.fetching[newHash] {
		t.Error("expected block to be marked as fetching")
	}

	done()
	if bs.fetching[newHash] {
		t.Error("expected block to be unmarked as fetching after cleanup")
	}
}
