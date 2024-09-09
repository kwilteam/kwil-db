package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// IPRateLimiter is a map of IP address to rate limiter.
type IPRateLimiter struct {
	ipsMtx sync.RWMutex
	ips    map[string]*rate.Limiter // ip address as key, rate limiter as value

	r     rate.Limit // refill rate, number of tokens per second
	burst int
}

func NewIPRateLimiter(rps float64, burst int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		r:     rate.Limit(rps),
		burst: burst,
	}

	return i
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) IP(ip string) *rate.Limiter {
	i.ipsMtx.Lock()
	defer i.ipsMtx.Unlock()

	limiter, exists := i.ips[ip]
	if exists {
		return limiter
	}
	limiter = rate.NewLimiter(i.r, i.burst)
	i.ips[ip] = limiter
	return limiter
}
