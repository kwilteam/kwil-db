package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync"

	"kwil/log"
	"kwil/node/types"

	"github.com/dgraph-io/badger/v4"
	boptions "github.com/dgraph-io/badger/v4/options"
)

// This version of BlockStore has one badger DB. The block is one value.
// Transaction gets seek into block.

type blockHashes struct {
	hash    types.Hash
	appHash types.Hash
}

type BlockStore struct {
	mtx      sync.RWMutex
	idx      map[types.Hash]int64
	hashes   map[int64]blockHashes
	fetching map[types.Hash]bool // TODO: remove, app concern

	// TODO: LRU cache for recent txns

	log log.Logger
	db  *badger.DB
}

var (
	nsHeader  = []byte("h:") // block metadata (header + signature)
	nsBlock   = []byte("b:") // full block
	nsTxn     = []byte("t:") // transaction index by tx hash
	nsAppHash = []byte("a:") // app hash by block hash
	nsResults = []byte("r:") // block execution results by block hash
)

func NewBlockStore(dir string, opts ...Option) (*BlockStore, error) {
	options := &options{
		logger: log.DiscardLogger,
	}

	for _, opt := range opts {
		opt(options)
	}
	logger := options.logger

	bOpts := badger.DefaultOptions(filepath.Join(dir, "bstore"))
	bOpts.WithLogger(&badgerLogger{logger})
	bOpts = bOpts.WithLoggingLevel(badger.WARNING)
	// bOpts.SyncWrites = true
	if options.compress {
		const bs = 1 << 16 // 18 = 256 KiB
		bOpts = bOpts.WithBlockSize(bs)
		bOpts.Compression = boptions.ZSTD
		bOpts.ZSTDCompressionLevel = 1
	} else {
		bOpts.Compression = boptions.None
	}
	bOpts = bOpts.WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(bOpts)
	if err != nil {
		return nil, err
	}
	bs := &BlockStore{
		idx:      make(map[types.Hash]int64),
		hashes:   make(map[int64]blockHashes),
		fetching: make(map[types.Hash]bool),
		db:       db,
		log:      logger,
	}

	// Initialize block index from the db
	itOpts := badger.DefaultIteratorOptions
	itOpts.Prefix = nsHeader
	pfxLen := len(nsHeader)

	var hash types.Hash // reuse in loop
	var appHash types.Hash

	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(itOpts)
		defer it.Close()

		var count int
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var height int64
			err := item.Value(func(val []byte) error {
				r := bytes.NewReader(val)
				blockHeader, err := types.DecodeBlockHeader(r)
				if err != nil {
					return err
				}
				height = blockHeader.Height
				// sig, err = io.ReadAll(r)
				return nil
			})
			if err != nil {
				return err
			}

			key := item.Key()
			if len(key) < types.HashLen+pfxLen {
				return errors.New("block hash in block index")
			}
			copy(hash[:], key[pfxLen:])

			// get the app hash from nsAppHash
			appHashKey := slices.Concat(nsAppHash, key[pfxLen:])
			item, err = txn.Get(appHashKey)
			if err != nil {
				return err
			}

			err = item.Value(func(val []byte) error {
				appHash = types.Hash(val)
				return nil
			})

			bs.idx[hash] = height
			bs.hashes[height] = blockHashes{
				hash:    hash,
				appHash: appHash,
			}
			count++
		}

		logger.Infof("indexed %d blocks", count)

		return nil
	})

	return bs, err
}

func (bki *BlockStore) Close() error {
	return bki.db.Close()
}

func (bki *BlockStore) Have(hash types.Hash) bool {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	_, have := bki.idx[hash]
	return have
}

const blkInfoLen = 8 + types.HashLen + 4

func makeTxVal(height int64, blkHash types.Hash, idx uint32) []byte {
	val := make([]byte, blkInfoLen)
	binary.LittleEndian.PutUint64(val, uint64(height))
	copy(val[8:], blkHash[:])
	binary.LittleEndian.PutUint32(val[8+types.HashLen:], idx)
	return val
}

func (bki *BlockStore) mayReplaceTx(txn *badger.Txn, err error) (*badger.Txn, error) {
	if err == nil {
		return txn, nil
	}
	if !errors.Is(err, badger.ErrTxnTooBig) {
		return nil, err
	}
	bki.log.Warn("block store: replacing large txn")
	if err = txn.Commit(); err != nil {
		txn.Discard()
		return nil, err
	}
	return bki.db.NewTransaction(true), nil
}

func (bki *BlockStore) StoreResults(hash types.Hash, results []types.TxResult) error {
	txn := bki.db.NewTransaction(true)
	defer txn.Discard()

	// The key prefix will be nsResults + tx index
	for i, res := range results {
		key := slices.Concat(nsResults, hash[:], binary.LittleEndian.AppendUint32(nil, uint32(i)))
		resBts, err := res.MarshalBinary()
		if err != nil {
			return err
		}
		err = txn.Set(key, resBts)
		if err != nil {
			txn, err = bki.mayReplaceTx(txn, err)
			if err != nil {
				return err
			} // else we recovered
			defer txn.Discard()
		}
	}

	return txn.Commit()
}

func (bki *BlockStore) Results(hash types.Hash) ([]types.TxResult, error) {
	prefixLen := len(nsResults) + types.HashLen

	// Get block header to determine number of transactions
	var txCount uint32
	err := bki.db.View(func(txn *badger.Txn) error {
		key := slices.Concat(nsHeader, hash[:])
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			header, err := types.DecodeBlockHeader(bytes.NewReader(val))
			if err != nil {
				return err
			}
			txCount = header.NumTxns
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	results := make([]types.TxResult, txCount)

	err = bki.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = slices.Concat(nsResults, hash[:])

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			idx := binary.LittleEndian.Uint32(key[prefixLen:])
			if idx >= txCount {
				return fmt.Errorf("invalid tx index %d", idx)
			}

			err := item.Value(func(val []byte) error {
				var result types.TxResult
				if err := result.UnmarshalBinary(val); err != nil {
					return err
				}
				results[idx] = result
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return results, err
}

func (bki *BlockStore) Store(blk *types.Block, appHash types.Hash) error {
	blkHash := blk.Hash()
	height := blk.Header.Height

	txn := bki.db.NewTransaction(true)
	defer txn.Discard()

	// Store block metadata (header + signature)
	key := slices.Concat(nsHeader, blkHash[:])
	err := txn.Set(key, append(types.EncodeBlockHeader(blk.Header), blk.Signature...))
	if err != nil {
		return err
	}

	// Store the block contents with the nsBlock prefix
	key = slices.Concat(nsBlock, blkHash[:])
	err = txn.Set(key, types.EncodeBlock(blk))
	if err != nil {
		return err
	}

	// Store the appHash.
	key = slices.Concat(nsAppHash, blkHash[:])
	err = txn.Set(key, appHash[:])
	if err != nil {
		return err
	}
	// NOTE: we are going to store this again in the next block's header, so
	// this is possibly a suboptimal design.

	// Store the txn index
	txIDs := make([]byte, 0, len(blk.Txns)*types.HashLen)
	for idx, tx := range blk.Txns {
		txHash := types.HashBytes(tx)
		key = slices.Concat(nsTxn, txHash[:]) // "t:txHash" => height + blkHash + blkIdx
		val := makeTxVal(height, blkHash, uint32(idx))
		err := txn.Set(key, val)
		txn, err = bki.mayReplaceTx(txn, err)
		if err != nil {
			return err
		}
		txIDs = append(txIDs, txHash[:]...)
	}

	bki.mtx.Lock()
	defer bki.mtx.Unlock()

	if err = txn.Commit(); err != nil {
		return err
	}

	delete(bki.fetching, blkHash)
	bki.idx[blkHash] = height
	bki.hashes[height] = blockHashes{
		hash:    blkHash,
		appHash: appHash,
	}

	return nil
}

func (bki *BlockStore) PreFetch(blkid types.Hash) (bool, func()) { // TODO: remove
	bki.mtx.Lock()
	defer bki.mtx.Unlock()
	if _, have := bki.idx[blkid]; have {
		return false, func() {} // don't need it
	}

	if fetching := bki.fetching[blkid]; fetching {
		return false, func() {} // already getting it
	}
	bki.fetching[blkid] = true

	return true, func() {
		bki.mtx.Lock()
		delete(bki.fetching, blkid)
		bki.mtx.Unlock()
	} // go get it
}

func (bki *BlockStore) size() int {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	return len(bki.idx)
}

func (bki *BlockStore) Best() (int64, types.Hash, types.Hash) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	var bestHeight int64
	var bestHash, bestAppHash types.Hash
	for height, hashes := range bki.hashes {
		if height >= bestHeight {
			bestHeight = height
			bestHash = hashes.hash
			bestAppHash = hashes.appHash
		}
	}
	return bestHeight, bestHash, bestAppHash
}

func (bki *BlockStore) Get(blkHash types.Hash) (*types.Block, types.Hash, error) {
	if !bki.Have(blkHash) {
		return nil, types.Hash{}, types.ErrNotFound
	}

	var block *types.Block
	var appHash types.Hash
	err := bki.db.View(func(txn *badger.Txn) error {
		// Load the block and get the tx
		blockKey := slices.Concat(nsBlock, blkHash[:])
		item, err := txn.Get(blockKey)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			block, err = types.DecodeBlock(val)
			return err
		})
		if err != nil {
			return err
		}

		// Load the apphash
		appHashKey := slices.Concat(nsAppHash, blkHash[:])
		item, err = txn.Get(appHashKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			appHash = types.Hash(val)
			return nil
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, types.Hash{}, types.ErrNotFound
	}

	return block, appHash, err
}

// GetByHeight retrieves the full block based on the block height. The returned
// hash is a convenience for the caller to spare computing it. The app hash,
// which is encoded in the next block header, is also returned.
func (bki *BlockStore) GetByHeight(height int64) (types.Hash, *types.Block, types.Hash, error) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()

	hashes, have := bki.hashes[height]
	if !have {
		return types.Hash{}, nil, types.Hash{}, types.ErrNotFound
	}
	blk, appHash, err := bki.Get(hashes.hash)
	if err != nil {
		return types.Hash{}, nil, types.Hash{}, err
	}
	// NOTE: appHash should match hashes.appHash, so check it here.
	if appHash != hashes.appHash {
		return types.Hash{}, nil, types.Hash{}, fmt.Errorf("appHash mismatch: %s != %s", appHash, hashes.appHash)
	}

	return hashes.hash, blk, appHash, err
}

func (bki *BlockStore) HaveTx(txHash types.Hash) bool {
	var have bool
	err := bki.db.View(func(txn *badger.Txn) error {
		key := slices.Concat(nsTxn, txHash[:]) // tdb["t:txHash"]
		if _, err := txn.Get(key); err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		have = true
		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrDBClosed) {
			return false
		}
		panic(err)
	}
	return have
}

func (bki *BlockStore) GetTx(txHash types.Hash) (int64, []byte, error) {
	var raw []byte
	var height int64
	var blkHash types.Hash
	var blkIdx uint32
	err := bki.db.View(func(txn *badger.Txn) error {
		// Get block info from the tx index
		key := slices.Concat(nsTxn, txHash[:]) // tdb["t:txHash"] => blk info
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			if len(val) < blkInfoLen {
				return errors.New("invalid block info for tx")
			}
			height = int64(binary.LittleEndian.Uint64(val))
			copy(blkHash[:], val[8:])
			blkIdx = binary.LittleEndian.Uint32(val[8+types.HashLen:])
			return nil
		})
		if err != nil {
			return err
		}

		// Load the block and get the tx
		blockKey := slices.Concat(nsBlock, blkHash[:])
		item, err = txn.Get(blockKey)
		if err != nil {
			return err // bug
		}

		return item.Value(func(val []byte) error {
			raw, err = types.GetRawBlockTx(val, blkIdx)
			return err
			// block, err := types.DecodeBlock(val)
			// if err != nil {
			// 	return err
			// }
			// if len(block.Txns) <= int(blkIdx) {
			// 	return types.ErrNotFound // bug
			// }
			// raw = block.Txns[blkIdx]
			// return nil
		})
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return 0, nil, types.ErrNotFound
		}
		return 0, nil, err
	}

	return height, raw, nil
}
