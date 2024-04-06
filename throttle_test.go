// Tideland Go Wait - Unit Tests
//
// Copyright (C) 2019-2023 Frank Mueller / Tideland / Oldenburg / Germany
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

	"tideland.dev/go/audit/asserts"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestThrottle verifies the throttling of parallel processed events.
func TestThrottle(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
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
		assert.Logf("test: %s", test.name)
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
			time.Sleep(25 * time.Millisecond)
			return nil
		}
		for i := 0; i < test.tasks; i++ {
			// Process the task in a goroutine.
			go func() {
				err := throttle.Process(ctx, task)
				wg.Done()
				if test.err == "" {
					assert.NoError(err)
				} else {
					assert.ErrorContains(err, test.err)
				}
			}()
		}
		wg.Wait()
		elapsed := time.Since(start)
		assert.Logf("elapsed: %v", elapsed)
		// Check the results.
		if test.burst > 0 {
			assert.Equal(cc.max(), test.burst, "maximum number of parallel goroutines defined by burst")
		}
		switch {
		case test.limit == 0 && test.burst == 0:
			assert.Equal(cc.max(), 0)
		case test.limit > 0 && test.burst == 1:
			expected := (time.Duration(test.tasks) / time.Duration(test.limit)) * time.Second
			tenth := expected / 10
			assert.Range(elapsed, expected-tenth, expected+tenth)
		}
	}
}

// TestThrottleBurst verifies the influence of the burst on throttling.
func TestThrottleBurst(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	results := [3][3]struct {
		burst   int
		tasks   int
		elapsed time.Duration
	}{}
	// Run nested tests.
	for i, burst := range []int{1, 5, 100} {
		for j, tasks := range []int{10, 50, 10000} {
			assert.Logf("burst: %d, tasks: %d", burst, tasks)
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
			for k := 0; k < tasks; k++ {
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
			assert.Logf("elapsed: %v", elapsed)
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
