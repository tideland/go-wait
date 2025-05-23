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

// TestPollWithJitter tests the polling with a jitter ticker of conditions.
func TestPollWithJitter(t *testing.T) {
	timestamps := []time.Time{}
	err := wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			500*time.Millisecond,
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

	for i := range 9 {
		diff := timestamps[i+1].Sub(timestamps[i])
		t.Logf("Diff %d: %v", i, diff)
		// According to implementation, jitter is within [offset, offset+interval]
		verify.InRange(t, diff, 10*time.Millisecond, 60*time.Millisecond)
	}
}

// TestJitter tests the convinience waiting with integrated jitter ticker.
func TestJitterWait(t *testing.T) {
	timestamps := []time.Time{}
	err := wait.WithJitter(
		context.Background(),
		50*time.Millisecond,
		10*time.Millisecond,
		500*time.Millisecond,
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			if len(timestamps) == 10 {
				return true, nil
			}
			return false, nil
		})
	verify.NoError(t, err)
	verify.Length(t, timestamps, 10)

	for i := range 9 {
		diff := timestamps[i+1].Sub(timestamps[i])
		t.Logf("Diff %d: %v", i, diff)
		// According to implementation, jitter is within [offset, offset+interval]
		verify.InRange(t, diff, 10*time.Millisecond, 60*time.Millisecond)
	}
}

// TestPollWithExceedingJitter tests if the jitter has a timeout before
// it signales the successful end.
func TestPollWithExceedingJitter(t *testing.T) {
	err := wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(
			50*time.Millisecond,
			10*time.Millisecond,
			500*time.Millisecond),
		func() (bool, error) {
			// Do anything consuming time.
			time.Sleep(50 * time.Millisecond)
			return false, nil
		},
	)
	verify.ErrorContains(t, err, "exceeded")
}

// EOF
