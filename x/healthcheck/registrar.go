package healthcheck

import (
	"context"
	"go.uber.org/zap"
	"kwil/x/logx"
	"time"
)

// Check represent a health check
type Check struct {
	Name    string
	Check   func(ctx context.Context) error
	Timeout time.Duration

	UpdateInterval time.Duration
	InitialDelay   time.Duration
}

// Registrar supports check registration and checker creation.
type Registrar interface {
	// RegisterCheck(check Check)   // register a sync check
	RegisterAsyncCheck(refreshPeriod time.Duration, initialDelay time.Duration, check Check)
	BuildChecker(checker Checker) Checker
}

type registrar struct {
	Checks []Check
	logger logx.Logger
}

func NewRegistrar() *registrar {
	return &registrar{logger: logx.New()}
}

func (r *registrar) RegisterAsyncCheck(refreshPeriod time.Duration, initialDelay time.Duration, check Check) {
	r.logger.Debug("Registering async check", zap.String("name", check.Name))

	check.UpdateInterval = refreshPeriod
	check.InitialDelay = initialDelay
	r.Checks = append(r.Checks, check)
}

func (r *registrar) BuildChecker(checker Checker) Checker {
	checker.Build(r.Checks)
	return checker
}
