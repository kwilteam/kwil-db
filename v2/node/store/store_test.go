package store

import (
	"bytes"
	"p2p/node/types"
	"testing"
	"time"
)

func setupTestBlockStore(t *testing.T) *blockStore {
	tmpDir := t.TempDir()
	bs, err := NewBlockStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create block store: %v", err)
	}

	t.Cleanup(func() {
		bs.db.Close()
		bs.tdb.Close()
	})

	return bs
}

func createTestBlock(height int64) *types.Block {
	txns := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
	}
	return types.NewBlock(height, types.Hash{2, 3, 4}, types.Hash{6, 7, 8},
		time.Unix(1729723553, 0), txns)
}

func TestBlockStore_StoreAndGet(t *testing.T) {
	bs := setupTestBlockStore(t)

	block := createTestBlock(1)
	bs.Store(block)

	hash := block.Hash()
	height, data := bs.Get(hash)

	if height != block.Header.Height {
		t.Errorf("Expected height %d, got %d", block.Header.Height, height)
	}

	if data == nil {
		t.Fatal("Expected block data, got nil")
	}

	retrievedBlock, err := types.DecodeBlock(data)
	if err != nil {
		t.Fatal(err)
	}
	if retrievedBlock.Header.Height != block.Header.Height {
		t.Errorf("Expected retrieved block height %d, got %d", block.Header.Height, retrievedBlock.Header.Height)
	}
	if retrievedBlock.Hash() != hash {
		t.Fatal("hash mismatch")
	}
}

func TestBlockStore_GetByHeight(t *testing.T) {
	bs := setupTestBlockStore(t)

	block := createTestBlock(1)
	bs.Store(block)

	hash, data := bs.GetByHeight(1)
	if hash != block.Hash() {
		t.Errorf("Expected hash %x, got %x", block.Hash(), hash)
	}

	if data == nil {
		t.Fatal("Expected block data, got nil")
	}
}

func TestBlockStore_Have(t *testing.T) {
	bs := setupTestBlockStore(t)

	block := createTestBlock(1)
	hash := block.Hash()

	if bs.Have(hash) {
		t.Error("Block should not exist before storing")
	}

	bs.Store(block)

	if !bs.Have(hash) {
		t.Error("Block should exist after storing")
	}
}

func TestBlockStore_GetTx(t *testing.T) {
	bs := setupTestBlockStore(t)

	block := createTestBlock(1)
	bs.Store(block)

	txHash := types.HashBytes(block.Txns[0])
	height, txData := bs.GetTx(txHash)

	if height != block.Header.Height {
		t.Errorf("Expected tx height %d, got %d", block.Header.Height, height)
	}

	if !bytes.Equal(txData, block.Txns[0]) {
		t.Error("Retrieved transaction data doesn't match original")
	}
}

func TestBlockStore_HaveTx(t *testing.T) {
	bs := setupTestBlockStore(t)

	block := createTestBlock(1)
	txHash := types.HashBytes(block.Txns[0])

	if bs.HaveTx(txHash) {
		t.Error("Transaction should not exist before storing block")
	}

	bs.Store(block)

	if !bs.HaveTx(txHash) {
		t.Error("Transaction should exist after storing block")
	}
}
