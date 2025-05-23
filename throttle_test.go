// Tideland Go Wait - Unit Tests
//
// Copyright (C) 2019-2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package wait_test // import "tideland.dev/go/wait"

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"sync"
	"testing"
	"time"

	"tideland.dev/go/asserts/verify"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestThrottle verifies the throttling of parallel processed events.
func TestThrottle(t *testing.T) {
	tests := []struct {
		name    string
		limit   wait.Limit
		burst   int
		tasks   int
		timeout time.Duration
		err     string
	}{
		{
			name:  "throttle allows no tasks",
			limit: 0,
			burst: 0,
			tasks: 10,
			err:   "Wait(n=1) exceeds limiter's burst 0",
		},
		{
			name:    "throttle has infinite limit and no burst",
			limit:   wait.InfLimit,
			burst:   0,
			tasks:   10,
			timeout: time.Second,
		},
		{
			name:  "throttle allows two tasks per second",
			limit: 2,
			burst: 1,
			tasks: 10,
		},
		{
			name:  "throttle allows five tasks per second",
			limit: 5,
			burst: 1,
			tasks: 10,
		},
	}
	// Run the different tests.
	for _, test := range tests {
		t.Logf("test: %s", test.name)
		throttle := wait.NewThrottle(test.limit, test.burst)
		ctx := context.Background()
		if test.timeout > 0 {
			// Add a timeout to the context.
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, test.timeout)
			defer cancel()
		}
		var wg sync.WaitGroup
		wg.Add(test.tasks)
		cc := &concurrencyCounter{}
		start := time.Now()
		task := func() error {
			cc.incr()
			defer cc.decr()
			// Each task takes a consistent amount of time
			time.Sleep(25 * time.Millisecond)
			return nil
		}
		for range test.tasks {
			// Process the task in a goroutine.
			go func() {
				err := throttle.Process(ctx, task)
				wg.Done()
				if test.err == "" {
					verify.NoError(t, err)
				} else {
					verify.ErrorContains(t, err, test.err)
				}
			}()
		}
		wg.Wait()
		elapsed := time.Since(start)
		t.Logf("elapsed: %v", elapsed)
		// Check the results.
		if test.burst > 0 {
			verify.Equal(t, cc.max(), test.burst, "maximum number of parallel goroutines defined by burst")
		}
		switch {
		case test.limit == 0 && test.burst == 0:
			verify.Equal(t, cc.max(), 0, "maximum number of parallel goroutines defined by burst")
		case test.limit > 0 && test.burst == 1:
			// For a throttle with limit N per second and burst 1:
			// - To process M tasks, it should take approximately M/N seconds
			// Don't try to be too precise with timing as it varies by system
			minTime := time.Duration(float64(test.tasks-1)/float64(test.limit)) * time.Second
			maxTime := time.Duration(float64(test.tasks+2)/float64(test.limit)) * time.Second
			verify.InRange(t, elapsed, minTime, maxTime, "elapsed time")
		}
	}
}

// TestThrottleBurst verifies the influence of the burst on throttling.
func TestThrottleBurst(t *testing.T) {
	results := [3][3]struct {
		burst   int
		tasks   int
		elapsed time.Duration
	}{}
	// Run nested tests.
	for i, burst := range []int{1, 5, 100} {
		for j, tasks := range []int{10, 50, 10000} {
			t.Logf("burst: %d, tasks: %d", burst, tasks)
			throttle := wait.NewThrottle(wait.InfLimit, burst)
			ctx := context.Background()
			var wg sync.WaitGroup
			wg.Add(tasks)
			cc := &concurrencyCounter{}
			start := time.Now()
			task := func() error {
				cc.incr()
				defer cc.decr()
				time.Sleep(25 * time.Millisecond)
				return nil
			}
			for k := range tasks {
				// Pause a bit every 250 tasks.
				if k%250 == 0 {
					time.Sleep(10 * time.Millisecond)
				}
				// Process the task in a goroutine.
				go func() {
					throttle.Process(ctx, task)
					wg.Done()
				}()
			}
			wg.Wait()
			elapsed := time.Since(start)
			t.Logf("elapsed: %v", elapsed)
			results[i][j].burst = burst
			results[i][j].tasks = tasks
			results[i][j].elapsed = elapsed
		}
	}
}

//--------------------
// HELPER
//--------------------

// concurrencyCounter is a helper to count the maximum number of
// parallel running goroutines.
type concurrencyCounter struct {
	mu      sync.Mutex
	current int
	maximum int
}

// increase increases the current number of goroutines and
// updates the maximum.
func (cc *concurrencyCounter) incr() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.current++
	if cc.current > cc.maximum {
		cc.maximum = cc.current
	}
}

// decrease decreases the current number of goroutines.
func (cc *concurrencyCounter) decr() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.current--
}

// maximum returns the maximum number of parallel running goroutines.
func (cc *concurrencyCounter) max() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	return cc.maximum
}

// EOF
