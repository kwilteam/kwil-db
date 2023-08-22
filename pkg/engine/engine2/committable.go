package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/sessions"
)

type EngineCommittable struct {
}

var _ sessions.Committable = (*EngineCommittable)(nil)

func (e *EngineCommittable) Apply(ctx context.Context, changes []byte) error {
	panic("TODO")
}

func (e *EngineCommittable) BeginApply(ctx context.Context) error {
	panic("TODO")
}

func (e *EngineCommittable) BeginCommit(ctx context.Context) error {
	panic("TODO")
}

func (e *EngineCommittable) Cancel(ctx context.Context) {
	panic("TODO")
}

func (e *EngineCommittable) EndApply(ctx context.Context) error {
	panic("TODO")
}

func (e *EngineCommittable) EndCommit(ctx context.Context, appender func([]byte) error) error {
	panic("TODO")
}

func (e *EngineCommittable) ID(ctx context.Context) ([]byte, error) {
	panic("TODO")
}
