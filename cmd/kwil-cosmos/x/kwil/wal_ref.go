package kwil

import (
	"sync/atomic"
	"unsafe"

	wal "github.com/kwilteam/kwil-db/internal/wal"
)

var emptyWal = unsafe.Pointer(&wal.Wal{})

type WalRef struct {
	log          *wal.Wal
	blockStarted *uint32
}

func (ref *WalRef) BeginBlock(ctx wal.BlockContext) {
	if uint32(86) == *ref.blockStarted {
		panic("the WAL has been closed.")
	}

	if !atomic.CompareAndSwapUint32(ref.blockStarted, 0, 1) {
		// essentially close it out since we need to error out
		ref.Close()
		panic("invalid operation. Wal already in begin block state.")
	}

	err := ref.log.BeginBlock(ctx)
	if err != nil {
		panic("unable to BeginBlock in WAL log. " + err.Error())
	}
}

func (ref *WalRef) EndBlock(ctx wal.BlockContext) {
	if uint32(86) == *ref.blockStarted {
		panic("the WAL has been closed.")
	}

	if !atomic.CompareAndSwapUint32(ref.blockStarted, 1, 2) {
		ref.Close()
		panic("the current WAL is NOT in a begin block state.")
	}

	err := ref.log.EndBlock(ctx)
	if err != nil {
		panic("unable to EndBlock in WAL log. " + err.Error())
	}

	if !atomic.CompareAndSwapUint32(ref.blockStarted, 2, 0) {
		ref.Close()
		panic("the current WAL is in an invalid state.")
	}
}

func (ref *WalRef) IsClosed() bool {
	return uint32(86) == *ref.blockStarted
}

func (ref *WalRef) Close() {
	if uint32(86) == *ref.blockStarted {
		return
	}

	*ref.blockStarted = 86

	ref.log.Close()
}

func Open(walFile string) *WalRef {
	v := uint32(0)
	return &WalRef{wal.Open(walFile), &v}
}
