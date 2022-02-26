// Tideland Go Wait - Unit Tests
//
// Copyright (C) 2019-2022 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package wait_test // import "tideland.dev/go/wait"

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"tideland.dev/go/audit/asserts"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestThrottleOK verifies the positiv throttling of parallel processed events.
func TestThrottleOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	runs := 110
	var rushes int64
	event := func() error {
		atomic.AddInt64(&rushes, 1)
		return nil
	}
	throttle := wait.NewThrottle(20.0, 1)
	rushing := func() {
		rush := func() {
			ctx := context.Background()
			throttle.Process(ctx, event)
		}
		for i := 0; i < runs; i++ {
			go rush()
		}
	}
	// All preparations done, now start rushing and check the load.
	start := time.Now()

	go rushing()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C

		rs := atomic.LoadInt64(&rushes)
		l := throttle.Limit()
		b := throttle.Burst()

		assert.Equal(l, wait.Limit(20.0))
		assert.Equal(b, 1)
		assert.Logf("RUN: %d / LIMIT: %.2f / BURST: %d", rs, l, b)

		if rs >= int64(runs) {
			break
		}
	}

	duration := time.Now().Sub(start)

	assert.Equal(rushes, int64(runs))
	assert.True(duration.Seconds() >= 5.0, "duration is", duration.String())
}

// TestThrottleBurstOK verifies the positiv throttling of parallel processed events with burst.
func TestThrottleBurstOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	runs := 110
	var rushes int64
	event := func() error {
		atomic.AddInt64(&rushes, 1)
		return nil
	}
	throttle := wait.NewThrottle(20.0, 5)
	rushing := func() {
		rush := func() {
			ctx := context.Background()
			throttle.Process(ctx, event, event, event, event, event)
		}
		for i := 0; i < runs/5; i++ {
			go rush()
		}
	}
	// All preparations done, now start rushing and check the load.
	start := time.Now()

	go rushing()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C

		rs := atomic.LoadInt64(&rushes)
		l := throttle.Limit()
		b := throttle.Burst()

		assert.Equal(l, wait.Limit(20.0))
		assert.Equal(b, 5)
		assert.Logf("RUN: %d / LIMIT: %.2f / BURST: %d", rs, l, b)

		if rs >= int64(runs) {
			break
		}
	}

	duration := time.Now().Sub(start)

	assert.Equal(rushes, int64(runs))
	assert.True(duration.Seconds() >= 5.0, "duration is", duration.String())
}

// EOF
