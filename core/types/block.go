package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/kwilteam/kwil-db/core/crypto"
)

const (
	BlockVersion = 0
)

var ErrNotFound = errors.New("not found")

type BlockHeader struct {
	Version  uint16
	Height   int64
	NumTxns  uint32
	PrevHash Hash // previous block's hash
	// app hash after last block.
	// calculated based on updates to the PG state, accounts, validators, chain state and txResults.
	PrevAppHash Hash
	Timestamp   time.Time
	MerkleRoot  Hash // Merkle tree reference to hash of all transactions for the block

	// Hash of the current validator set for the block
	ValidatorSetHash Hash

	// ConsensusParams updates for the block, empty if no updates, ignored for this release
	// ConsensusParamsUpdates *ConsensusParams
}

type Block struct {
	Header    *BlockHeader
	Txns      []*Transaction
	Signature []byte // Signature is the block producer's signature (leader in our model)
}

func NewBlock(height int64, prevHash, appHash, valSetHash Hash, stamp time.Time, txns []*Transaction) *Block {
	numTxns := len(txns)
	txHashes := make([]Hash, numTxns)
	for i, tx := range txns {
		txHashes[i] = tx.Hash()
	}
	merkelRoot := CalcMerkleRoot(txHashes)
	hdr := &BlockHeader{
		Version:     BlockVersion,
		Height:      height,
		PrevHash:    prevHash,
		PrevAppHash: appHash,
		NumTxns:     uint32(numTxns),
		Timestamp:   stamp.UTC(),
		MerkleRoot:  merkelRoot,

		ValidatorSetHash: valSetHash,
	}
	return &Block{
		Header: hdr,
		Txns:   txns,
	}
}

func (b *Block) Hash() Hash {
	return b.Header.Hash()
}

func (b *Block) MerkleRoot() Hash {
	txHashes := make([]Hash, len(b.Txns))
	for i, tx := range b.Txns {
		txHashes[i] = tx.Hash()
	}
	return CalcMerkleRoot(txHashes)
}

func (b *Block) Sign(signer crypto.PrivateKey) error {
	hash := b.Hash()
	sig, err := signer.Sign(hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign block: %w", err)
	}
	b.Signature = sig
	return nil
}

func (b *Block) VerifySignature(pubKey crypto.PublicKey) (bool, error) {
	hash := b.Hash()
	return pubKey.Verify(hash[:], b.Signature)
}

func DecodeBlockHeader(r io.Reader) (*BlockHeader, error) {
	hdr := new(BlockHeader)

	if err := binary.Read(r, binary.LittleEndian, &hdr.Version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hdr.Height); err != nil {
		return nil, fmt.Errorf("failed to read height: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &hdr.NumTxns); err != nil {
		return nil, fmt.Errorf("failed to read number of transactions: %w", err)
	}

	_, err := io.ReadFull(r, hdr.PrevHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read previous block hash: %w", err)
	}

	_, err = io.ReadFull(r, hdr.PrevAppHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read previous block hash: %w", err)
	}

	_, err = io.ReadFull(r, hdr.ValidatorSetHash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read validator hash: %w", err)
	}

	var timeStamp uint64
	if err := binary.Read(r, binary.LittleEndian, &timeStamp); err != nil {
		return nil, fmt.Errorf("failed to read number of transactions: %w", err)
	}
	hdr.Timestamp = time.UnixMilli(int64(timeStamp)).UTC()

	_, err = io.ReadFull(r, hdr.MerkleRoot[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read merkel root: %w", err)
	}

	// Read validator updates

	return hdr, nil
}

func (bh *BlockHeader) String() string {
	return fmt.Sprintf("BlockHeader{Version: %d, Height: %d, NumTxns: %d, PrevHash: %s, AppHash: %s, Timestamp: %s, MerkelRoot: %s}",
		bh.Version,
		bh.Height,
		bh.NumTxns,
		bh.PrevHash,
		bh.PrevAppHash,
		bh.Timestamp.Format(time.RFC3339),
		bh.MerkleRoot)
}

func (bh *BlockHeader) writeBlockHeader(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, bh.Version); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	if err := binary.Write(w, binary.LittleEndian, bh.Height); err != nil {
		return fmt.Errorf("failed to write height: %w", err)
	}

	if err := binary.Write(w, binary.LittleEndian, bh.NumTxns); err != nil {
		return fmt.Errorf("failed to write number of transactions: %w", err)
	}

	if _, err := w.Write(bh.PrevHash[:]); err != nil {
		return fmt.Errorf("failed to write previous block hash: %w", err)
	}

	if _, err := w.Write(bh.PrevAppHash[:]); err != nil {
		return fmt.Errorf("failed to write app hash: %w", err)
	}

	if _, err := w.Write(bh.ValidatorSetHash[:]); err != nil {
		return fmt.Errorf("failed to write validator hash: %w", err)
	}

	if err := binary.Write(w, binary.LittleEndian, uint64(bh.Timestamp.UnixMilli())); err != nil {
		return fmt.Errorf("failed to write timestamp: %w", err)
	}

	if _, err := w.Write(bh.MerkleRoot[:]); err != nil {
		return fmt.Errorf("failed to write merkel root: %w", err)
	}

	// for _, v := range bh.ValidatorUpdates {
	// 	if err := binary.Write(w, binary.LittleEndian, v.PubKey); err != nil {
	// 		return fmt.Errorf("failed to write validator pubkey: %w", err)
	// 	}

	// 	if err := binary.Write(w, binary.LittleEndian, v.Power); err != nil {
	// 		return fmt.Errorf("failed to write validator power: %w", err)
	// 	}
	// }

	return nil
}

func EncodeBlockHeader(hdr *BlockHeader) []byte {
	var buf bytes.Buffer
	hdr.writeBlockHeader(&buf)
	return buf.Bytes()
}

func (bh *BlockHeader) Hash() Hash {
	hasher := NewHasher()
	bh.writeBlockHeader(hasher)
	return hasher.Sum(nil)
}

/*func encodeBlockHeaderOneAlloc(hdr *BlockHeader) []byte {
	// Preallocate buffer: 2 + 8 + 4 + 32 + 32 + 8 + 32 = 118 bytes
	buf := make([]byte, 0, 118)

	buf = binary.LittleEndian.AppendUint16(buf, hdr.Version)
	buf = binary.LittleEndian.AppendUint64(buf, uint64(hdr.Height))
	buf = binary.LittleEndian.AppendUint32(buf, hdr.NumTxns)
	buf = append(buf, hdr.PrevHash[:]...)
	buf = append(buf, hdr.AppHash[:]...)
	buf = binary.LittleEndian.AppendUint64(buf, uint64(hdr.Timestamp.UnixMilli()))
	buf = append(buf, hdr.MerkelRoot[:]...)

	return buf
}*/

func EncodeBlock(block *Block) []byte {
	headerBytes := EncodeBlockHeader(block.Header)

	buf := make([]byte, 0, len(headerBytes)+4+len(block.Signature)) // it's a lot more depending on txns, but we don't have size functions yet

	buf = append(buf, headerBytes...)

	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(block.Signature)))
	buf = append(buf, block.Signature...)

	for _, tx := range block.Txns {
		rawTx := tx.Bytes()
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(rawTx)))
		buf = append(buf, rawTx...)
	}

	return buf
}

// CalcMerkleRoot computes the merkel root for a slice of hashes. This is based
// on the "inline" implementation from btcd / dcrd.
func CalcMerkleRoot(leaves []Hash) Hash {
	switch len(leaves) {
	case 0:
		return Hash{}
	case 1:
		return leaves[0]
	default:
	}

	// Do not modify the leaves slice from the caller.
	leaves = slices.Clone(leaves)

	// Create a buffer to reuse for hashing the branches and some long lived
	// slices into it to avoid reslicing.
	var buf [2 * HashLen]byte
	var left = buf[:HashLen]
	var right = buf[HashLen:]
	var both = buf[:]

	// The following algorithm works by replacing the leftmost entries in the
	// slice with the concatenations of each subsequent set of 2 hashes and
	// shrinking the slice by half to account for the fact that each level of
	// the tree is half the size of the previous one.  In the case a level is
	// unbalanced (there is no final right child), the final node is duplicated
	// so it ultimately is concatenated with itself.
	//
	// For example, the following illustrates calculating a tree with 5 leaves:
	//
	// [0 1 2 3 4]                              (5 entries)
	// 1st iteration: [h(0||1) h(2||3) h(4||4)] (3 entries)
	// 2nd iteration: [h(h01||h23) h(h44||h44)] (2 entries)
	// 3rd iteration: [h(h0123||h4444)]         (1 entry)
	for len(leaves) > 1 {
		// When there is no right child, the parent is generated by hashing the
		// concatenation of the left child with itself.
		if len(leaves)&1 != 0 {
			leaves = append(leaves, leaves[len(leaves)-1])
		}

		// Set the parent node to the hash of the concatenation of the left and
		// right children.
		for i := range len(leaves) / 2 {
			copy(left, leaves[i*2][:])
			copy(right, leaves[i*2+1][:])
			leaves[i] = HashBytes(both)
		}
		leaves = leaves[:len(leaves)/2]
	}
	return leaves[0]
}

func DecodeBlock(rawBlk []byte) (*Block, error) {
	r := bytes.NewReader(rawBlk)

	hdr, err := DecodeBlockHeader(r)
	if err != nil {
		return nil, err
	}

	var sigLen uint32
	if err := binary.Read(r, binary.LittleEndian, &sigLen); err != nil {
		return nil, fmt.Errorf("failed to read signature length: %w", err)
	}

	if int(sigLen) > r.Len() { // more than remaining
		return nil, fmt.Errorf("invalid signature length %d", sigLen)
	}

	sig := make([]byte, sigLen)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	txns := make([]*Transaction, hdr.NumTxns)

	for i := range txns {
		var txLen uint32
		if err := binary.Read(r, binary.LittleEndian, &txLen); err != nil {
			return nil, fmt.Errorf("failed to read tx length: %w", err)
		}

		if int(txLen) > r.Len() { // more than remaining
			return nil, fmt.Errorf("invalid transaction length %d", txLen)
		}

		rawTx := make([]byte, txLen)
		if _, err := io.ReadFull(r, rawTx); err != nil {
			return nil, fmt.Errorf("failed to read tx data: %w", err)
		}
		var tx Transaction
		if err = tx.UnmarshalBinary(rawTx); err != nil {
			return nil, fmt.Errorf("invalid transaction (%d): %w", i, err)
		}
		txns[i] = &tx
	}

	return &Block{
		Header:    hdr,
		Txns:      txns,
		Signature: sig,
	}, nil
}

// GetRawBlockTx extracts a transaction from a serialized block by its index in
// the block. This allows to more efficiently extract one transaction without
// copying all of the transactions in the block, and it avoids hashing all of
// the transactions, which would be required to match by txID.
func GetRawBlockTx(rawBlk []byte, idx uint32) ([]byte, error) {
	r := bytes.NewReader(rawBlk)

	hdr, err := DecodeBlockHeader(r)
	if err != nil {
		return nil, err
	}

	var sigLen uint32
	if err := binary.Read(r, binary.LittleEndian, &sigLen); err != nil {
		return nil, fmt.Errorf("failed to read signature length: %w", err)
	}

	if int(sigLen) > len(rawBlk) { // TODO: do better than this
		return nil, fmt.Errorf("invalid signature length %d", sigLen)
	}

	sig := make([]byte, sigLen)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	for i := range hdr.NumTxns {
		var txLen uint32
		if err := binary.Read(r, binary.LittleEndian, &txLen); err != nil {
			return nil, fmt.Errorf("failed to read tx length: %w", err)
		}
		if int(txLen) > len(rawBlk) { // TODO: do better than this
			return nil, fmt.Errorf("invalid transaction length %d", txLen)
		}
		if idx == i {
			tx := make([]byte, txLen)
			if _, err := io.ReadFull(r, tx); err != nil {
				return nil, fmt.Errorf("failed to read tx data: %w", err)
			}
			return tx, nil
		}
		// seek to the start of the next tx
		if _, err := r.Seek(int64(txLen), io.SeekCurrent); err != nil {
			return nil, fmt.Errorf("failed to seek to next tx: %w", err)
		}
	}
	return nil, ErrNotFound
}

/*func decodeBlockTxns(raw []byte) (txns [][]byte, err error) {
	rd := bytes.NewReader(raw)
	for rd.Len() > 0 {
		var txLen uint64
		if err := binary.Read(rd, binary.LittleEndian, &txLen); err != nil {
			return nil, fmt.Errorf("failed to read tx length: %w", err)
		}

		if txLen > uint64(rd.Len()) {
			return nil, fmt.Errorf("invalid tx length %d", txLen)
		}
		tx := make([]byte, txLen)
		if _, err := io.ReadFull(rd, tx); err != nil {
			return nil, fmt.Errorf("failed to read tx data: %w", err)
		}
		txns = append(txns, tx)
	}

	return txns, nil
}*/
