package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
)

const (
	BlockVersion = 0
)

var HashBytes = types.HashBytes

const HashLen = types.HashLen

type Hash = types.Hash
type BlockHeader = types.BlockHeader
type Block = types.Block

func NewBlock(height int64, prevHash, appHash, valSetHash Hash, stamp time.Time, txns [][]byte) *Block {
	numTxns := len(txns)
	txHashes := make([]Hash, numTxns)
	for i := range txns {
		txHashes[i] = HashBytes(txns[i])
	}
	merkelRoot := types.CalcMerkleRoot(txHashes)
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

func EncodeBlockHeader(hdr *BlockHeader) []byte {
	var buf bytes.Buffer
	hdr.WriteBlockHeader(&buf)
	return buf.Bytes()
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

	totalSize := len(headerBytes)
	for _, tx := range block.Txns {
		totalSize += 4 + len(tx) // 4 bytes for transaction length
	}

	totalSize += 4 + len(block.Signature) // 4 bytes for signature length

	buf := make([]byte, 0, totalSize)

	buf = append(buf, headerBytes...)

	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(block.Signature)))
	buf = append(buf, block.Signature...)

	for _, tx := range block.Txns {
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(tx)))
		buf = append(buf, tx...)
	}

	return buf
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

	if int(sigLen) > len(rawBlk) { // TODO: do better than this
		return nil, fmt.Errorf("invalid signature length %d", sigLen)
	}

	sig := make([]byte, sigLen)
	if _, err := io.ReadFull(r, sig); err != nil {
		return nil, fmt.Errorf("failed to read signature: %w", err)
	}

	txns := make([][]byte, hdr.NumTxns)

	for i := range txns {
		var txLen uint32
		if err := binary.Read(r, binary.LittleEndian, &txLen); err != nil {
			return nil, fmt.Errorf("failed to read tx length: %w", err)
		}

		if int(txLen) > len(rawBlk) { // TODO: do better than this
			return nil, fmt.Errorf("invalid transaction length %d", txLen)
		}

		tx := make([]byte, txLen)
		if _, err := io.ReadFull(r, tx); err != nil {
			return nil, fmt.Errorf("failed to read tx data: %w", err)
		}
		txns[i] = tx
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
