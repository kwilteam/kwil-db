package retry

import (
	"context"
	"kwil/pkg/log"
	"time"
)

type RetryOpt func(*retrier)

func WithLogger(log log.Logger) RetryOpt {
	return func(r *retrier) {
		r.log = log
	}
}

func WithFactor(factor float64) RetryOpt {
	return func(r *retrier) {
		r.retrier.Factor = factor
	}
}

func WithoutJitter() RetryOpt {
	return func(r *retrier) {
		r.retrier.Jitter = false
	}
}

func WithMax(max time.Duration) RetryOpt {
	return func(r *retrier) {
		r.retrier.Max = max
	}
}

func WithMin(min time.Duration) RetryOpt {
	return func(r *retrier) {
		r.retrier.Min = min
	}
}

func WithMaxRetries(maxRetries int) RetryOpt {
	return func(r *retrier) {
		r.maxRetries = maxRetries
	}
}

func WithContext(ctx context.Context) RetryOpt {
	return func(r *retrier) {
		r.ctx = ctx
	}
}
