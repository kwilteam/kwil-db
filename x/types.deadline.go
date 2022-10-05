package x

import "time"

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
	return !d.deadline.Before(time.Now())
}

// NewDeadline constructs a new Deadline with the given timeout.
func NewDeadline(timeout time.Duration) *Deadline {
	return &Deadline{time.Now().Add(time.Millisecond * timeout)}
}
