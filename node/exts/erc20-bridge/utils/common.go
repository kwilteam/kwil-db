package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/jpillora/backoff"
)

// Retry will retry the function until it is successful, or reached the max retries
func Retry(ctx context.Context, maxRetries int64, fn func() error) error {
	retrier := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	for {
		err := fn()
		if err == nil {
			return nil
		}

		// fail after maxRetries retries
		if retrier.Attempt() > float64(maxRetries) {
			return fmt.Errorf("retry maxout: %w", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retrier.Duration()):
		}
	}
}
