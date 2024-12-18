package main

import (
	"context"
	"errors"
	"strings"
	"time"

	lorem "github.com/drhodes/golorem"
)

type printer func(msg string, args ...any)

func makeBigData(maxLen int, print printer) string {
	var bigDataBuilder strings.Builder
	var sentences int
	for bigDataBuilder.Len() < maxLen {
		words := min(16, contentLen/16+1)
		bigDataBuilder.WriteString(lorem.Sentence(words, words)) // this may be far considerably than needed because words vary in length, but that's fine
		bigDataBuilder.WriteString(" ")
		sentences++
	}
	print("Generated content from %d lorem ipsum sentences.", sentences)
	return bigDataBuilder.String()
}

// asyncFn wraps some asynchronous function, gating with the provided semaphore,
// ultimately discarding the result, and logging errors and timing info.
func asyncFn(ctx context.Context, sem chan struct{}, print printer, operation string,
	fn func() (<-chan asyncResp, error)) func() error {
	return func() error {
		sem <- struct{}{}

		t0 := time.Now()
		defer func() {
			since := time.Since(t0)
			var slow string
			if since > 200*time.Millisecond {
				slow = " (SLOW)"
			}
			print("%s took %vms%s", operation, float64(since.Microseconds())/1e3, slow)
		}()

		promise, err := fn()
		if err != nil {
			// Continue runLooped if it was a timeout as these are typically
			// transient and we don't want to let up.
			if errors.Is(err, context.DeadlineExceeded) {
				print("%s timeout", operation)
				err = nil // just keep trying
			}
			<-sem
			return err
		}

		go func() {
			timer := time.NewTimer(13 * time.Second)
			defer timer.Stop()
			select {
			case res := <-promise:
				if err := res.Error(); err != nil {
					print("%s failed: %v", operation, err)
				}
			case <-timer.C:
				print("timed out waiting for %s result (e.g. tx to be mined)", operation)
			case <-ctx.Done():
			}

			<-sem
		}()
		return nil
	}
}
