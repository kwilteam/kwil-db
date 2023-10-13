package healthcheck

import (
	"context"
)

type Result struct {
	Status string
}

type Checker interface {
	Start()
	Stop()
	Check(ctx context.Context) Result
	Build([]Check)
}
