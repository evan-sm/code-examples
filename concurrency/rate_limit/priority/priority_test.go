package priority

import (
	"fmt"
	"math"
	"testing"
	"testing/synctest"
	"time"
)

const chatID int64 = 42

func TestPriorityLimiterAllow(t *testing.T) {
	// t.Parallel()

	type step struct {
		sleep      time.Duration
		priority   Priority
		wantAllow  bool
		wantTokens float64
	}

	tests := []struct {
		name     string
		rate     float64
		capacity float64
		reserve  float64
		steps    []step
	}{
		{
			name:     "burst & refill",
			rate:     1,
			capacity: 10,
			reserve:  6,
			steps: []step{
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 9},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 8},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 7},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 6},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 5},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: false, wantTokens: 5},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: false, wantTokens: 5},
				{sleep: 0 * time.Second, priority: PriorityLow, wantAllow: false, wantTokens: 5},
				{sleep: 2 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 6},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 5},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 4},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 3},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 2},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 1},
				{sleep: 0 * time.Second, priority: PriorityHigh, wantAllow: true, wantTokens: 0},
				{sleep: 10 * time.Second, priority: PriorityLow, wantAllow: true, wantTokens: 9},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				limiter := New(tt.rate, tt.capacity, tt.reserve)

				for i, s := range tt.steps {
					fmt.Println("step", i)
					time.Sleep(s.sleep)
					synctest.Wait()

					got := limiter.Allow(chatID, s.priority, time.Now())
					if got != s.wantAllow {
						t.Fatalf("step %d priority %v: Allow = %v, want %v", i, s.priority, got, s.wantAllow)
					}

					b := limiter.buckets[chatID]
					if b == nil {
						t.Fatalf("step %d: bucket missing", i)
					}
					if diff := math.Abs(b.tokens - s.wantTokens); diff > 1e-9 {
						t.Fatalf("step %d priority %v: tokens = %.2f, want %.2f (diff %.2g)", i, s.priority, b.tokens, s.wantTokens, diff)
					}
				}
			})

		})
	}
}
