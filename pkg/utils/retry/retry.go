package retry

import (
	"context"
	"fmt"
	"kwil/pkg/log"
	"time"

	"github.com/jpillora/backoff"
	"go.uber.org/zap"
)

type Retrier[T any] struct {
	val        T
	retrier    *backoff.Backoff
	log        log.Logger
	maxRetries int
}

type retryMethod[T any] func(context.Context, T) error

// New creates a new Retrier[T] with the given value and options.
// Golang cannot infer the type of the options, so this must be declared specifically.
// Example:
//
//	retrier := retry.New(strct,
//		retry.WithFactor[*TestStruct](2),
//		retry.WithMax[*TestStruct](time.Millisecond*5000),
//	)
func New[T any](val T, opts ...RetryOpt[T]) *Retrier[T] {
	r := &Retrier[T]{
		val: val,
		retrier: &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    10 * time.Second,
			Factor: 2,
			Jitter: true,
		},
		log:        log.NewNoOp(),
		maxRetries: -1,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Retrier[T]) Retry(ctx context.Context, method retryMethod[T]) error {
	counter := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := method(ctx, r.val)
			if err == nil {
				r.retrier.Reset()
				return nil
			}

			counter++
			r.log.Error("retrier: error occurred, retrying", zap.Error(err), zap.Int("retry_count", counter))
			if r.exceedsMaxRetries(counter) {
				r.retrier.Reset()
				return fmt.Errorf("retrier: exceeded max retries (%d)", r.maxRetries)
			}

			time.Sleep(r.retrier.Duration())
		}
	}
}

func (r *Retrier[T]) exceedsMaxRetries(retryNumber int) bool {
	if r.maxRetries == -1 {
		return false
	}
	return retryNumber > r.maxRetries
}
