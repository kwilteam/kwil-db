package peers

import (
	"math"
	"math/rand"
	"time"
)

type backoffer struct {
	maxAttempts int
	baseDelay   time.Duration
	maxDelay    time.Duration
	jitter      bool

	attempts int
	last     time.Time // could just be next really...
}

// newBackoffer creates a new exponential backoff helper. The first attempts
// delay is 0. The first try() always returns true, or the first next() always
// return the Duration 0. This type's methods are not safe for concurrent use.
func newBackoffer(maxAttempts int, baseDelay, maxDelay time.Duration, jitter bool) *backoffer {
	return &backoffer{
		maxAttempts: maxAttempts,
		baseDelay:   baseDelay,
		maxDelay:    maxDelay,
		jitter:      jitter,
	}
}

// goTime is a helper function to return the next time to try a request. It does
// not increment the attempt counter. Instead, use try() or next(), which both
// increment the attempt counter.
func (b *backoffer) goTime() time.Time {
	var jitter time.Duration
	if b.jitter {
		jitter = time.Duration(rand.Int63n(int64(b.baseDelay)))
	}
	delay := min(b.maxDelay, jitter+b.baseDelay*(1<<(b.attempts-1)))
	return b.last.Add(delay)
}

// try returns true if an attempt should be made. This only increments the
// attempt counter if it returns true. The first call always returns true.
//
// Use this method when making attempts from a pool of multiple resources, and
// searching for a resource that is ready. Use next() and maxedOut() when
// waiting on a single resource in a loop.
func (b *backoffer) try() bool {
	now := time.Now()
	if b.attempts == 0 {
		b.last = now
		b.attempts++
		return true
	}
	if b.attempts >= b.maxAttempts {
		return false
	}

	okTime := b.goTime()

	okToTry := now.After(okTime)
	if okToTry {
		b.attempts++
		b.last = now
	}
	return okToTry
}

// next gives the time for the next attempt. Calling this is an attempt (it is
// increments the attempt counter). This should be used in a loop with
// `time.After`, in conjunction with maxedOut() to break the loop when the
// maximum attempts have been reached (or track attempts in the loop).
func (b *backoffer) next() time.Duration {
	// this is an attempt
	b.last = time.Now()
	b.attempts++

	if b.attempts == 1 {
		return 0
	}

	okTime := b.goTime()

	return max(0, time.Until(okTime))
}

func (b *backoffer) tries() int {
	return b.attempts
}

// maxedOut returns true if the maximum number of attempts has been reached.
func (b *backoffer) maxedOut() bool {
	return b.attempts >= b.maxAttempts
}

// calculateBackoffTTL computes total backoff time with jitter for n retries
// (base 2). This is used when adding a peer to the peerstore so that the TTL
// used by libp2p for removing the peer matches our own connect retry logic with
// exponential backoff.
// 2sec base, 1hr max, 58 retries should be close to 48 hrs
func calculateBackoffTTL(baseDelay, maxDelay time.Duration, retries int, jitter bool) time.Duration {
	var totalBackoff time.Duration

	// 63 - log2 of basedelay is the max exp before overflow
	maxExp := int(63 - math.Ceil(math.Log2(float64(baseDelay))))

	for i := range retries {
		delay := baseDelay * (1 << min(i, maxExp)) // baseDelay * 2^i
		if delay < 0 {                             // catch overflow
			return time.Duration(math.MaxInt64)
		}

		// Add average jitter: baseDelay / 2 to approximate the average impact.
		delayWithJitter := delay
		if jitter {
			avgJitter := baseDelay / 2
			delayWithJitter = delay + avgJitter
		}

		cappedDelay := min(delayWithJitter, maxDelay)
		totalBackoff += cappedDelay
		if totalBackoff < 0 { // catch overflow
			return time.Duration(math.MaxInt64)
		}
	}

	return totalBackoff
}
