package blockchain

import (
	"context"

	"github.com/google/uuid"
)

// The current blockTx and message propagation will be
// done using the same JSON format as a passthrough.
// This will be replaced with a more efficient and appropriate
// schema/format in the future.
type BlockTxStatus uint
type BlockTxId []byte

type BlockTx struct {
	id    BlockTxId
	group []byte
	data  []byte
}

func (b *BlockTx) GetId() BlockTxId {
	return b.id
}

const (
	Pending BlockTxStatus = iota
	Complete
	Failed
	Cancelled
)

func CreateBlockTx(group []byte, data []byte) BlockTx {
	u := uuid.New()
	return BlockTx{[]byte(u.String()), group, data}
}

type ChainTxCallback struct {
	ctx ChainContext
	fn  func(ChainContext, error)
}

func (c *ChainTxCallback) Error(err error) {
	c.fn(c.ctx, err)
}

func (c *ChainTxCallback) Success() {
	c.fn(c.ctx, nil)
}

// A Chain processes up to 10,000 transactions per second.
type Chain interface {
	// Unique chain name
	GetName() string

	// Submits a new block to the *network*.
	Submit(ctx context.Context, tx *BlockTx, fn *ChainTxCallback)

	// Sets the handler for this named chain. This can only be set prior to chain start.
	SetHandler(handler ChainTxHandler) //Using a single instance for now
}

type ChainContext interface {
	context.Context

	// Gets the chain by name.
	ChainId() string
	GetHeight() uint64

	tx() *BlockTx
	opaque() interface{}
}

type ChainTxHandler interface {
	// Called when a BlockTx is received.
	// Validate(ctx *ChainContext, tx *BlockTx) error

	// Called for each tx.
	CommitBlock(fn *ChainTxCallback)
	Close()
}
