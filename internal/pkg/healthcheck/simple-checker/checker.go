package simple_checker

import (
	"context"

	healthcheck2 "github.com/kwilteam/kwil-db/internal/pkg/healthcheck"
	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/alexliesenfeld/health"
	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var statusMap = map[string]string{
	string(health.StatusUp):      grpc_health_v1.HealthCheckResponse_SERVING.String(),
	string(health.StatusDown):    grpc_health_v1.HealthCheckResponse_NOT_SERVING.String(),
	string(health.StatusUnknown): grpc_health_v1.HealthCheckResponse_UNKNOWN.String(),
}

type SimpleChecker struct {
	Ck     health.Checker
	logger log.Logger
}

func New(logger log.Logger) *SimpleChecker {
	return &SimpleChecker{logger: *logger.Named("healthcheck.simple-checker")}
}

func (c *SimpleChecker) Start() {
	c.Ck.Start()
}

func (c *SimpleChecker) Stop() {
	c.Ck.Stop()
}

func (c *SimpleChecker) Check(ctx context.Context) healthcheck2.Result {
	res := c.Ck.Check(ctx)
	return healthcheck2.Result{Status: statusMap[string(res.Status)]}
}

func (c *SimpleChecker) Build(checks []healthcheck2.Check) {
	var cks []health.CheckerOption
	for _, ck := range checks {
		if ck.UpdateInterval > 0 {
			cks = append(cks, health.WithPeriodicCheck(ck.UpdateInterval, ck.InitialDelay, health.Check{
				Name:  ck.Name,
				Check: ck.Check,
			}))
		} else {
			cks = append(cks, health.WithCheck(health.Check{
				Name:  ck.Name,
				Check: ck.Check,
			}))
		}
	}

	cks = append(cks,
		health.WithStatusListener(func(ctx context.Context, state health.CheckerState) {
			c.logger.Info("Health check state changed", zap.String("state", string(state.Status)))
		}))

	c.Ck = health.NewChecker(cks...)
}
