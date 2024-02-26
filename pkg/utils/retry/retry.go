package retry

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/log"
	"time"

	"github.com/jpillora/backoff"
	"go.uber.org/zap"
)

type retrier struct {
	retrier    *backoff.Backoff
	log        log.Logger
	maxRetries int
	ctx        context.Context
}

func Retry(fn func() error, opts ...RetryOpt) error {
	ret := &retrier{
		retrier: &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    10 * time.Second,
			Factor: 2,
			Jitter: true,
		},
		log:        log.NewNoOp(),
		maxRetries: -1,
		ctx:        context.Background(),
	}

	for _, opt := range opts {
		opt(ret)
	}

	counter := 0
	for {
		select {
		case <-ret.ctx.Done():
			return ret.ctx.Err()
		default:
			err := fn()
			if err == nil {
				ret.retrier.Reset()
				return nil
			}

			counter++
			ret.log.Error("retrier: error occurred, retrying", zap.Error(err), zap.Int("retry_count", counter))
			if ret.exceedsMaxRetries(counter) {
				ret.retrier.Reset()
				return fmt.Errorf("retrier: exceeded max retries (%d)", ret.maxRetries)
			}

			time.Sleep(ret.retrier.Duration())
		}
	}
}

func (ret *retrier) exceedsMaxRetries(retryNumber int) bool {
	if ret.maxRetries == -1 {
		return false
	}
	return retryNumber > ret.maxRetries
}
