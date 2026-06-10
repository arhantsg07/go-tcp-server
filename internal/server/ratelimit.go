package server

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket algorithm per connection.
// It allows short bursts up to `capacity` messages, then sustains `rate` messages/second.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	rate       float64 // tokens added per second
	lastRefill time.Time
}

// NewRateLimiter creates a token bucket with the given burst capacity and refill rate.
func NewRateLimiter(capacity, rate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Allow returns true if this request is within the rate limit.
// It refills the bucket based on elapsed time before deciding.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens = min(r.capacity, r.tokens+elapsed*r.rate)
	r.lastRefill = now

	if r.tokens < 1 {
		return false
	}
	r.tokens--
	return true
}
