package retry

import (
	"kwil/pkg/log"
	"time"
)

type RetryOpt[T any] func(*Retrier[T])

func WithLogger[T any](log log.Logger) RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.log = log
	}
}

func WithFactor[T any](factor float64) RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.retrier.Factor = factor
	}
}

func WithoutJitter[T any]() RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.retrier.Jitter = false
	}
}

func WithMax[T any](max time.Duration) RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.retrier.Max = max
	}
}

func WithMin[T any](min time.Duration) RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.retrier.Min = min
	}
}

func WithMaxRetries[T any](maxRetries int) RetryOpt[T] {
	return func(r *Retrier[T]) {
		r.maxRetries = maxRetries
	}
}
