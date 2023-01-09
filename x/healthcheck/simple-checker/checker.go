package simple_checker

import (
	"context"
	"github.com/alexliesenfeld/health"
	"go.uber.org/zap"
	"kwil/x/healthcheck"
	"kwil/x/logx"
	"kwil/x/proto/healthpb"
)

var statusMap = map[string]string{
	string(health.StatusUp):      healthpb.HealthCheckResponse_SERVING.String(),
	string(health.StatusDown):    healthpb.HealthCheckResponse_NOT_SERVING.String(),
	string(health.StatusUnknown): healthpb.HealthCheckResponse_UNKNOWN.String(),
}

type SimpleChecker struct {
	Ck     health.Checker
	logger logx.Logger
}

func New() *SimpleChecker {
	return &SimpleChecker{logger: logx.New()}
}

func (c *SimpleChecker) Start() {
	c.Ck.Start()
}

func (c *SimpleChecker) Stop() {
	c.Ck.Stop()
}

func (c *SimpleChecker) Check(ctx context.Context) healthcheck.Result {
	res := c.Ck.Check(ctx)
	return healthcheck.Result{Status: statusMap[string(res.Status)]}
}

func (c *SimpleChecker) Build(checks []healthcheck.Check) {
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
