package dt

import (
	"time"
)

// Deadline implements the deadline/timeout resiliency pattern.
type Deadline struct {
	deadline time.Time
}

func (d *Deadline) Expiry() time.Time {
	return d.deadline
}

func (d *Deadline) HasExpired() bool {
	return !d.deadline.Before(time.Now())
}

// NewDeadline constructs a new Deadline with the given timeout.
func NewDeadline(timeoutMillis time.Duration) *Deadline {
	return &Deadline{time.Now().Add(time.Millisecond * timeoutMillis)}
}
