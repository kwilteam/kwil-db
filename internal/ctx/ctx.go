package ctx

import (
	"github.com/kwilteam/kwil-db/internal/wal"
	"golang.org/x/net/context"
	"time"
)

type KwilCtxFactory interface {
	CreateCtxFactory(parent context.Context) KwilContext
}

type KwilContext interface {
	context.Context
	Wal() *wal.WalDbCmd
}

type kwilContextImpl struct {
	parent context.Context
	walRef *wal.WalDbCmd
}

func NewKwilContext(parent context.Context, walRef *wal.WalDbCmd) KwilContext {
	return &kwilContextImpl{parent, walRef}
}

func (k *kwilContextImpl) Wal() *wal.WalDbCmd { return k.walRef }

const kwilContextKey string = "kwil-context"

func Unwrap(ctx context.Context) KwilContext {
	return ctx.Value(kwilContextKey).(KwilContext)
}

func (k *kwilContextImpl) Deadline() (deadline time.Time, ok bool) {
	return k.parent.Deadline()
}

func (k *kwilContextImpl) Done() <-chan struct{} {
	return k.parent.Done()
}

func (k *kwilContextImpl) Err() error {
	return k.parent.Err()
}

func (k *kwilContextImpl) Value(key any) any {
	if key == kwilContextKey {
		return k
	}
	return k.parent.Value(key)
}
