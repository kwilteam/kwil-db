package async

import (
	"context"
	"kwil/x"
)

type WorkerPool interface {
	Execute(job x.Job)
	ExecuteWith(ctx context.Context, job x.Job)
	Submit(job x.Job) Action

	Start() error

	IsRunning() bool
	IsShutdown() bool
	Shutdown() error
	OnShutdown() <-chan x.Void

	IsShutdownRequested() bool
	OnShutdownRequested() <-chan x.Void
}
