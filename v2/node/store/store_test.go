package store

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"p2p/node/types"
)

func getFileSizes(dirPath string) ([][2]string, error) {
	var filesInfo [][2]string

	// Walk through the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only get files, not directories
		if !info.IsDir() {
			// Store file name and size in KB
			filesInfo = append(filesInfo, [2]string{info.Name(), fmt.Sprintf("%.2f KiB", float64(info.Size())/1024)})
		}
		return nil
	})
	return filesInfo, err
}

// Pretty print function
func prettyPrintFileSizes(filesInfo [][2]string) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(writer, "File Name\tSize")
	fmt.Fprintln(writer, "---------\t----")

	for _, fileInfo := range filesInfo {
		fmt.Fprintf(writer, "%s\t%s\n", fileInfo[0], fileInfo[1])
	}
	writer.Flush()
}

func setupTestBlockStore(t *testing.T, compress ...bool) (*BlockStore, string) {
	comp := len(compress) > 0 && compress[0]
	tmpDir := t.TempDir()
	bs, err := NewBlockStore(tmpDir, WithCompression(comp))
	if err != nil {
		t.Fatalf("Failed to create block store: %v", err)
	}

	t.Cleanup(func() {
		bs.Close()
	})

	return bs, tmpDir
}

func createTestBlock(height int64, numTxns int) *types.Block {
	txns := make([][]byte, numTxns)
	for i := 0; i < numTxns; i++ {
		txns[i] = []byte(strconv.FormatInt(height, 10) + strconv.Itoa(i) +
			strings.Repeat("data", 1000))
	}
	return types.NewBlock(height, types.Hash{2, 3, 4}, types.Hash{6, 7, 8},
		time.Unix(1729723553+height, 0), txns)
}

func TestBlockStore_StoreAndGet(t *testing.T) {
	bs, _ := setupTestBlockStore(t)

	block := createTestBlock(1, 2)
	bs.Store(block)

	hash := block.Hash()
	blk, err := bs.Get(hash)
	if err != nil {
		t.Fatal(err)
	}
	height, data := blk.Header.Height, types.EncodeBlock(blk)

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
	bs, _ := setupTestBlockStore(t)

	block := createTestBlock(1, 2)
	bs.Store(block)

	gotHash, blk, err := bs.GetByHeight(1)
	if err != nil {
		t.Fatal(err)
	}
	hash := blk.Hash()
	if hash != block.Hash() {
		t.Errorf("Expected hash %x, got %x", block.Hash(), hash)
	}
	if hash != gotHash {
		t.Errorf("Expected hash %x, got %x", block.Hash(), hash)
	}

	// if data == nil {
	// 	t.Fatal("Expected block data, got nil")
	// }
}

func TestBlockStore_Have(t *testing.T) {
	bs, _ := setupTestBlockStore(t)

	block := createTestBlock(1, 2)
	hash := block.Hash()

	if bs.Have(hash) {
		t.Error("Block should not exist before storing")
	}

	bs.Store(block)

	if !bs.Have(hash) {
		t.Error("Block should exist after storing")
	}

	bs.Close()
}

func TestBlockStore_GetTx(t *testing.T) {
	bs, _ := setupTestBlockStore(t)

	block := createTestBlock(1, 3)
	bs.Store(block)

	for i := range block.Txns {
		txHash := types.HashBytes(block.Txns[i])
		height, txData, err := bs.GetTx(txHash)
		if err != nil {
			t.Fatal(err)
		}

		if height != block.Header.Height {
			t.Errorf("Expected tx height %d, got %d", block.Header.Height, height)
		}

		if !bytes.Equal(txData, block.Txns[i]) {
			t.Error("Retrieved transaction data doesn't match original")
		}
	}
}

func TestBlockStore_HaveTx(t *testing.T) {
	bs, dir := setupTestBlockStore(t)

	block := createTestBlock(1, 6)
	txHash := types.HashBytes(block.Txns[0])

	if bs.HaveTx(txHash) {
		t.Error("Transaction should not exist before storing block")
	}

	bs.Store(block)

	if !bs.HaveTx(txHash) {
		t.Error("Transaction should exist after storing block")
	}

	bs.Close()

	bs, err := NewBlockStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if !bs.HaveTx(txHash) {
		t.Error("Transaction should exist after reloading store")
	}

	if !bs.Have(block.Hash()) {
		t.Error("block should exist after reloading store")
	}
}

func TestBlockStore_StoreWithNoTransactions(t *testing.T) {
	bs, _ := setupTestBlockStore(t)
	block := types.NewBlock(1, types.Hash{2, 3, 4}, types.Hash{6, 7, 8},
		time.Unix(1729723553, 0), [][]byte{})

	err := bs.Store(block)
	if err != nil {
		t.Fatal(err)
	}

	blk, err := bs.Get(block.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if len(blk.Txns) != 0 {
		t.Error("Expected empty transactions")
	}
}

func TestBlockStore_StoreWithEmptyTransactions(t *testing.T) {
	bs, _ := setupTestBlockStore(t)
	block := types.NewBlock(1, types.Hash{2, 3, 4}, types.Hash{6, 7, 8},
		time.Unix(1729723553, 0), [][]byte{{}, {}})

	err := bs.Store(block)
	if err != nil {
		t.Fatal(err)
	}

	blk, err := bs.Get(block.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if len(blk.Txns) != 2 {
		t.Error("Expected two transactions")
	}
}

func TestBlockStore_StoreConcurrent(t *testing.T) {
	bs, _ := setupTestBlockStore(t)
	done := make(chan bool)
	blockCount := 100

	for i := 0; i < 3; i++ {
		go func(start int) {
			for j := 0; j < blockCount; j++ {
				block := createTestBlock(int64(start*blockCount+j), 2)
				err := bs.Store(block)
				if err != nil {
					t.Error(err)
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < 3; i++ {
		<-done
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < blockCount; j++ {
			height := int64(i*blockCount + j)
			_, blk, err := bs.GetByHeight(height)
			if err != nil {
				t.Errorf("Failed to get block at height %d: %v", height, err)
			}
			if blk.Header.Height != height {
				t.Errorf("Expected height %d, got %d", height, blk.Header.Height)
			}
		}
	}
}

func TestBlockStore_StoreDuplicateBlock(t *testing.T) {
	bs, _ := setupTestBlockStore(t)
	block := createTestBlock(1, 2)

	err := bs.Store(block)
	if err != nil {
		t.Fatal(err)
	}

	err = bs.Store(block)
	if err != nil {
		t.Fatal(err)
	}

	height, hash := bs.Best()
	if height != block.Header.Height {
		t.Errorf("Expected height %d, got %d", block.Header.Height, height)
	}
	if hash != block.Hash() {
		t.Errorf("Expected hash %x, got %x", block.Hash(), hash)
	}
}

func TestBlockStore_StoreWithLargeTransactions(t *testing.T) {
	bs, _ := setupTestBlockStore(t, true)
	largeTx := make([]byte, 1<<20) // 1MB transaction
	for i := range largeTx {
		largeTx[i] = byte(i % 256)
	}
	otherTx := []byte{1, 2, 3}

	block := types.NewBlock(1, types.Hash{2, 3, 4}, types.Hash{6, 7, 8},
		time.Unix(1729723553, 0), [][]byte{largeTx, otherTx})

	err := bs.Store(block)
	if err != nil {
		t.Fatal(err)
	}

	blkHash := block.Hash()
	bk, err := bs.Get(blkHash)
	if err != nil {
		t.Fatal(err)
	}
	if bk.Hash() != blkHash {
		t.Fatal("hash mismatch")
	}

	for _, tx := range [][]byte{largeTx, otherTx} {
		txHash := types.HashBytes(tx)
		_, txData, err := bs.GetTx(txHash)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(txData, tx) {
			t.Error("Retrieved transaction data doesn't match original")
		}
	}

	// bs.Close()
	// filesInfo, err := getFileSizes(tmpDir)
	// if err != nil {
	// 	t.Logf("Error: %v", err)
	// } else {
	// 	prettyPrintFileSizes(filesInfo)
	// }
}

func TestLargeBlockStore(t *testing.T) {
	// This test demonstrates that zstd level 1 compression is faster than no
	// compression for reasonably compressible data.

	// Create block store
	dir := t.TempDir()
	bs, err := NewBlockStore(dir, WithCompression(true))
	if err != nil {
		t.Fatal(err)
	}
	defer bs.Close()

	// Generate large number of blocks with many txs
	const numBlocks = 100
	const txsPerBlock = 1000
	const txSize = 1024 // 1KB per tx, small enough not to use vlog for the block

	var prevHash types.Hash
	var prevAppHash types.Hash

	rngSrc := rand.NewChaCha8(prevHash)
	rng := rand.New(rngSrc)

	// Patterned tx body to make it compressible
	txBody := make([]byte, txSize-8)
	for i := range txBody {
		txBody[i] = byte(i % 16)
	}

	// Create blocks with random transactions
	for height := int64(1); height <= numBlocks; height++ {
		// Generate random transactions
		txs := make([][]byte, txsPerBlock)
		for i := range txs {
			tx := make([]byte, txSize)
			rngSrc.Read(tx[:8]) // like a nonce, ensures txs are unique
			copy(tx[8:], txBody)
			txs[i] = tx
		}

		// Create and store block
		block := types.NewBlock(height, prevHash, prevAppHash, time.Now(), txs)
		err = bs.Store(block)
		if err != nil {
			t.Fatal(err)
		}

		prevHash = block.Hash()
		prevAppHash = types.HashBytes([]byte(fmt.Sprintf("app-%d", height)))

		// Verify block retrieval
		if height%100 == 0 {
			retrieved, err := bs.Get(prevHash)
			if err != nil {
				t.Errorf("Failed to get block at height %d: %v", height, err)
			}
			if retrieved.Hash() != prevHash {
				t.Errorf("Retrieved block hash mismatch at height %d", height)
			}

			// Verify random transaction retrieval

			txIdx := rng.IntN(len(txs))
			txHash := types.HashBytes(txs[txIdx])
			gotHeight, gotTx, err := bs.GetTx(txHash)
			if err != nil {
				t.Errorf("Failed to get tx at height %d, idx %d: %v", height, txIdx, err)
			}
			if gotHeight != height {
				t.Errorf("Wrong tx height. Got %d, want %d", gotHeight, height)
			}
			if !bytes.Equal(gotTx, txs[txIdx]) {
				t.Error("Retrieved tx data mismatch")
			}
		}
	}

	// Verify final database size
	bs.Close() // else we are checking size of huge sparse allocated files
	dbSize := getDirSize(dir)
	t.Logf("Final database size: %d MB", dbSize/1024/1024)
	filesInfo, err := getFileSizes(dir)
	if err != nil {
		t.Logf("Error: %v", err)
	} else {
		prettyPrintFileSizes(filesInfo)
	}
}

func getDirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
