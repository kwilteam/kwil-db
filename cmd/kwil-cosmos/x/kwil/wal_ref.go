package kwil

import (
	"sync/atomic"
	"unsafe"

	"github.com/kwilteam/kwil-db/internal/wal"
)

var emptyWal = unsafe.Pointer(&wal.Wal{})

type WalRef struct {
	log *wal.Wal
}

func CreateWalRef() *WalRef {
	return &WalRef{log: (*wal.Wal)(emptyWal)}
}

func (ref *WalRef) BeginBlock(ctx wal.BlockContext) {
	nextWal, err := wal.NewBlockWal(ctx)
	if err != nil {
		panic("unable to create new WAL log. " + err.Error())
	}

	p := (*unsafe.Pointer)(unsafe.Pointer(&ref.log))
	if atomic.CompareAndSwapPointer(p, emptyWal, unsafe.Pointer(&nextWal)) {
		return
	}

	// essentially close it out since we need to error out
	nextWal.Seal()

	panic("the current WAL has not been disposed.")
}

func (ref *WalRef) EndBlock() {
	log := ref.log
	p := (*unsafe.Pointer)(unsafe.Pointer(&ref.log))
	if !atomic.CompareAndSwapPointer(p, unsafe.Pointer(ref.log), emptyWal) {
		panic("the current WAL has been modified or sealed out of sequence with the block begin/end events.")
	}

	err := log.Seal()
	if err != nil {
		panic("unable to seal WAL log. " + err.Error())
	}
}
