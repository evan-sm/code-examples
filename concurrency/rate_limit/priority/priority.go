// Package priority shows how to implement a priority-based rate limiter.
package priority

import (
	"sync"
	"time"
)

type Priority int

const (
	PriorityHigh Priority = iota
	PriorityLow
)

// PriorityLimiter holds tokens.
type PriorityLimiter struct {
	mu       sync.RWMutex
	rate     float64 // tokens per second
	capacity float64 // max tokens per bucket (burst)
	reserve  float64 // tokens to keep for high priority

	buckets map[int64]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func New(rate, capacity, reserve float64) *PriorityLimiter {
	if reserve > capacity {
		reserve = capacity
	}

	return &PriorityLimiter{
		rate:     rate,
		capacity: capacity,
		reserve:  reserve,
		buckets:  make(map[int64]*bucket),
	}
}

// Allow checks if a request is allowed for the given chat ID and priority.
func (l *PriorityLimiter) Allow(chatID int64, priority Priority, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[chatID]
	if !ok {
		b = &bucket{
			tokens: l.capacity, // start full capacity
			last:   now,
		}
		l.buckets[chatID] = b
	}

	// Update the bucket's tokens based on the time elapsed since the last request.
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity { // cannot exceed capacity
		b.tokens = l.capacity
	}
	b.last = now

	// Allow based on the priority.
	switch priority {
	case PriorityHigh:
		if b.tokens < 1 {
			return false
		}

		b.tokens--
		return true
	case PriorityLow:
		if b.tokens < l.reserve {
			return false
		}

		b.tokens--
		return true
	}

	return false
}
