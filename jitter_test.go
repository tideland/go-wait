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
	"testing"
	"time"

	"tideland.dev/go/asserts/verify"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestPollWithJitter tests the polling of conditions in a maximum
// number of intervals.
func TestPollWithJitter(t *testing.T) {
	timestamps := []time.Time{}
	err := wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			1250*time.Millisecond,
		),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			if len(timestamps) == 10 {
				return true, nil
			}
			return false, nil
		},
	)
	verify.NoError(t, err)
	verify.Length(t, timestamps, 10)
	
	t.Logf("Timestamps for first test: %v", timestamps)
	for i := range 9 {
		diff := timestamps[i+1].Sub(timestamps[i])
		t.Logf("Diff %d: %v", i, diff)
		// According to implementation, jitter is within [offset, offset+interval]
		verify.InRange(t, 10*time.Millisecond, 60*time.Millisecond, diff)
	}

	timestamps = []time.Time{}
	err = wait.WithJitter(
		context.Background(),
		50*time.Millisecond,
		10*time.Millisecond,
		1250*time.Millisecond, func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			if len(timestamps) == 10 {
				return true, nil
			}
			return false, nil
		})
	verify.NoError(t, err)
	verify.Length(t, timestamps, 10)
	
	t.Logf("Timestamps for second test: %v", timestamps)
	for i := 1; i < 10; i++ {
		diff := timestamps[i].Sub(timestamps[i-1])
		t.Logf("Diff %d: %v", i, diff)
		// According to implementation, jitter is within [offset, offset+interval]
		verify.InRange(t, 10*time.Millisecond, 60*time.Millisecond, diff)
	}

	timestamps = []time.Time{}
	err = wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			1250*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	verify.ErrorContains(t, "exceeded", err)
	verify.InRange(t, len(timestamps), 10, 25)

	timestamps = []time.Time{}
	err = wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			1000*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	verify.ErrorContains(t, "exceeded", err)
	// This ticker has a non-zero offset, so it should run at least once before timing out
	verify.True(t, len(timestamps) > 0)

	timestamps = []time.Time{}
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	err = wait.Poll(
		ctx,
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			1000*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	verify.ErrorContains(t, "exceeded", err)
	// Context cancellation should still allow some ticks to happen
	verify.True(t, len(timestamps) > 0)
}

// EOF