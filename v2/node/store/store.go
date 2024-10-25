package store

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sync"

	"p2p/log"
	"p2p/node/types"

	"github.com/dgraph-io/badger/v4"
)

// This version of BlockStore has one badger DB,

type BlockStore struct {
	mtx      sync.RWMutex
	idx      map[types.Hash]int64
	hashes   map[int64]types.Hash // []types.Hash, also could store pointer
	fetching map[types.Hash]bool  // TODO: remove, app concern

	log log.Logger
	db  *badger.DB
}

var (
	nsHeader = []byte("h:") // block metadata (header + signature)
	nsBlock  = []byte("b:") // full block
	nsTxn    = []byte("t:") // transaction index by tx hash
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
	// opts.SyncWrites = true
	// opts.Compression = options.ZSTD
	// opts.ZSTDCompressionLevel = 3
	bOpts = bOpts.WithLoggingLevel(badger.WARNING)
	db, err := badger.Open(bOpts)
	if err != nil {
		return nil, err
	}
	bs := &BlockStore{
		idx:      make(map[types.Hash]int64),
		hashes:   make(map[int64]types.Hash),
		fetching: make(map[types.Hash]bool),
		db:       db,
		log:      logger,
	}

	// Initialize block index from the db
	itOpts := badger.DefaultIteratorOptions
	itOpts.Prefix = nsHeader
	pfxLen := len(nsHeader)

	var hash types.Hash // reuse in loop

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
			bs.idx[hash] = height
			bs.hashes[height] = hash
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

func (bki *BlockStore) Store(blk *types.Block) error {
	blkHash := blk.Hash()
	height := blk.Header.Height

	bki.mtx.Lock()
	defer bki.mtx.Unlock()
	delete(bki.fetching, blkHash)
	bki.idx[blkHash] = height
	bki.hashes[height] = blkHash

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

	return txn.Commit()
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

func (bki *BlockStore) Best() (int64, types.Hash) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()
	var bestHeight int64
	var bestHash types.Hash
	for height, hash := range bki.hashes {
		if height >= bestHeight {
			bestHeight = height
			bestHash = hash
		}
	}
	return bestHeight, bestHash
}

func (bki *BlockStore) Get(blkHash types.Hash) (*types.Block, error) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()

	var block *types.Block
	err := bki.db.View(func(txn *badger.Txn) error {
		// Load the block and get the tx
		blockKey := slices.Concat(nsBlock, blkHash[:])
		item, err := txn.Get(blockKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			block, err = types.DecodeBlock(val)
			return err
		})
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, types.ErrNotFound
	}

	return block, err
}

// GetByHeight retrieves the full block based on the block height. The returned
// hash is a convenience for the caller to spare computing it.
func (bki *BlockStore) GetByHeight(height int64) (types.Hash, *types.Block, error) {
	bki.mtx.RLock()
	defer bki.mtx.RUnlock()

	hash, have := bki.hashes[height]
	if !have {
		return types.Hash{}, nil, fmt.Errorf("block not found at height %d", height)
	}
	blk, err := bki.Get(hash)
	return hash, blk, err
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
