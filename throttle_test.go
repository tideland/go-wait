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
	"tideland.dev/go/audit/generators"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestThrottleOK verifies the positiv limitation of parallel processed events.
func TestThrottleOK(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	gen := generators.New(generators.FixedRand())
	runs := 110
	event := func() error {
		gen.SleepOneOf(10*time.Millisecond, 20*time.Millisecond, 50*time.Millisecond)
		return nil
	}
	throttle := wait.NewThrottle(20)
	var rushes int64
	rushing := func() {
		rush := func() {
			ctx := context.Background()
			throttle.Process(ctx, event)
			atomic.AddInt64(&rushes, 1)
		}
		for i := 0; i < runs; i++ {
			go rush()
		}
	}
	// All preparations done, now start rushing and check the load.
	before := time.Now()

	go rushing()

	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C

		rs := atomic.LoadInt64(&rushes)
		ll := throttle.LimitedLoad()
		cl := throttle.CurrentLoad()

		assert.True(cl >= 0 && cl < 20)
		assert.Logf("RUN: %d / LL: %d / CL: %d", rs, ll, cl)

		if rs >= int64(runs) {
			break
		}
	}

	after := time.Now()
	distance := after.Sub(before)

	assert.True(distance.Seconds() > 5.0)
}

// EOF
