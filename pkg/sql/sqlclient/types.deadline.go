package sqlclient

import (
	"time"
)

// MaxTimeoutAllowed is the maximum timeout allowed for
// a deadline (1 hour).
const MaxTimeoutAllowed = time.Duration(1 * time.Hour)

// Deadline implements the deadline/timeout resiliency pattern.
type Deadline struct {
	deadline time.Time
}

// Expiry will return the time at which the deadline will expire
func (d *Deadline) Expiry() time.Time {
	return d.deadline
}

// HasExpired will return true if the deadline has expired
func (d *Deadline) HasExpired() bool {
	return d.deadline.Before(time.Now())
}

// Remaining will return the remaining duration before the deadline expires.
func (d *Deadline) Remaining() time.Duration {
	remaining := d.deadline.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (d *Deadline) RemainingMillis() int {
	remaining := d.Remaining()
	if remaining == 0 {
		return 0
	}

	return int(remaining.Milliseconds())
}

// NewDeadline constructs a new Deadline with the given timeout.
// A panic will result if timeout > MaxTimeoutAllowed. A negative
// timeout will result in a deadline that has already expired.
func NewDeadline(timeout time.Duration) *Deadline {
	if timeout > MaxTimeoutAllowed {
		panic("timeout is greater than the maximum allowed")
	}

	if timeout < 0 {
		timeout = -1
	}

	return &Deadline{time.Now().Add(timeout)}
}
