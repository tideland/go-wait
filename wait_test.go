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

// TestPolls verifies Poll() with different parameters.
func TestPolls(t *testing.T) {
	tests := []struct {
		name              string
		ticker            func() wait.TickerFunc
		duration          time.Duration
		expectedCount     int
		expectedErrorText string
	}{
		{
			name: "interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeIntervalTicker(5 * time.Millisecond)
			},
			expectedCount: 5,
		}, {
			name: "interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeIntervalTicker(5 * time.Millisecond)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name:          "changing-interval-poll-success",
			ticker:        mkChgTicker,
			expectedCount: 5,
		}, {
			name:              "changing-interval-poll-ticker-exceeds",
			ticker:            mkChgTicker,
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name:              "changing-interval-poll-context-cancelled",
			ticker:            mkChgTicker,
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "max-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			expectedCount: 5,
		}, {
			name: "max-interval-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name: "max-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "deadlined-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			expectedCount: 5,
		}, {
			name: "deadlined-interval-poll-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name: "deadline-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "expiring-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			expectedCount: 5,
		}, {
			name: "expiring-interval-poll-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name: "expiring-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "expiring-max-intervals-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringMaxIntervalsTicker(5*time.Millisecond, 55*time.Millisecond, 10)
			},
			expectedCount: 5,
		}, {
			name: "expiring-max-intervals-poll-max-exceeded",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringMaxIntervalsTicker(5*time.Millisecond, 100*time.Millisecond, 3)
			},
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name: "expiring-max-intervals-poll-timeout-exceeded",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringMaxIntervalsTicker(5*time.Millisecond, 25*time.Millisecond, 10)
			},
			expectedErrorText: "ticker exceeded while waiting for the condition",
		}, {
			name: "expiring-max-intervals-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringMaxIntervalsTicker(5*time.Millisecond, 55*time.Millisecond, 10)
			},
			duration:          20 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		},
	}
	// Run tests.
	ct := verify.ContinuedTesting(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.duration != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, test.duration)
				verify.NotNil(ct, cancel)
			}
			count := 0
			condition := func() (bool, error) {
				count++
				if count == test.expectedCount {
					return true, nil
				}
				return false, nil
			}
			err := wait.Poll(ctx, test.ticker(), condition)
			if test.expectedErrorText == "" {
				verify.NoError(t, err)
				verify.Equal(t, count, test.expectedCount)
			} else {
				verify.ErrorContains(t, err, test.expectedErrorText)
			}
		})
	}
}

// TestConvenience verifies the diverse convenience functions for Poll().
func TestConvenience(t *testing.T) {
	tests := []struct {
		name              string
		duration          time.Duration
		poll              func(context.Context, wait.ConditionFunc) error
		expectedCount     int
		expectedErrorText string
	}{
		{
			name: "with-interval-success",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithInterval(ctx, 5*time.Millisecond, condition)
			},
			expectedCount: 5,
		}, {
			name: "with-interval-success-context-cancelled",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithInterval(ctx, 5*time.Millisecond, condition)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "with-max-intervals-success",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithMaxIntervals(ctx, 5*time.Millisecond, 10, condition)
			},
			expectedCount: 5,
		}, {
			name: "with-max-intervals-context-cancelled",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithMaxIntervals(ctx, 5*time.Millisecond, 10, condition)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "with-deadline-success",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithDeadline(ctx, 5*time.Millisecond, time.Now().Add(55*time.Millisecond), condition)
			},
			expectedCount: 5,
		}, {
			name: "with-deadline-context-cancelled",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithDeadline(ctx, 5*time.Millisecond, time.Now().Add(55*time.Millisecond), condition)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		}, {
			name: "with-timeout-success",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithTimeout(ctx, 5*time.Millisecond, 55*time.Millisecond, condition)
			},
			expectedCount: 5,
		}, {
			name: "with-timeout-context-cancelled",
			poll: func(ctx context.Context, condition wait.ConditionFunc) error {
				return wait.WithTimeout(ctx, 5*time.Millisecond, 55*time.Millisecond, condition)
			},
			duration:          50 * time.Millisecond,
			expectedErrorText: "context has been cancelled with error",
		},
	}
	// Run tests.
	ct := verify.ContinuedTesting(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.duration != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, test.duration)
				verify.NotNil(ct, cancel)
			}
			count := 0
			condition := func() (bool, error) {
				count++
				if count == test.expectedCount {
					return true, nil
				}
				return false, nil
			}
			err := test.poll(ctx, condition)
			if test.expectedErrorText == "" {
				verify.NoError(ct, err)
				verify.Equal(ct, count, test.expectedCount)
			} else {
				verify.ErrorContains(ct, err, test.expectedErrorText)
			}
		})
	}
}

// TestExpiringMaxIntervalsTicker tests the combination of maximum intervals and timeout.
func TestExpiringMaxIntervalsTicker(t *testing.T) {
	// Test cases for our ticker
	tests := []struct {
		name          string
		interval      time.Duration
		timeout       time.Duration
		maxIntervals  int
		condition     func(int) bool // Returns true when count should trigger success
		expectError   bool
		expectedTicks int // How many ticks we expect before either success or timeout
	}{
		{
			name:          "success-before-max-intervals",
			interval:      10 * time.Millisecond,
			timeout:       500 * time.Millisecond,
			maxIntervals:  10,
			condition:     func(count int) bool { return count == 5 },
			expectError:   false,
			expectedTicks: 5,
		},
		{
			name:          "max-intervals-reached-first",
			interval:      10 * time.Millisecond,
			timeout:       500 * time.Millisecond,
			maxIntervals:  5,
			condition:     func(count int) bool { return count > 10 }, // Never reaches this
			expectError:   true,
			expectedTicks: 5,
		},
		{
			name:          "timeout-reached-first",
			interval:      50 * time.Millisecond,
			timeout:       100 * time.Millisecond,
			maxIntervals:  10,
			condition:     func(count int) bool { return count > 10 }, // Never reaches this
			expectError:   true,
			expectedTicks: 2, // Only have time for about 2 ticks
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tickCount := 0
			ticker := wait.MakeExpiringMaxIntervalsTicker(
				test.interval,
				test.timeout,
				test.maxIntervals,
			)

			err := wait.Poll(
				context.Background(),
				ticker,
				func() (bool, error) {
					tickCount++
					return test.condition(tickCount), nil
				},
			)

			if test.expectError {
				verify.ErrorContains(t, err, "exceeded")
			} else {
				verify.NoError(t, err)
			}

			verify.Equal(t, tickCount, test.expectedTicks)
		})
	}
}

// TestUserDefinedTicker tests the polling of conditions with a user-defined ticker.
func TestUserDefinedTicker(t *testing.T) {
	ticker := func(ctx context.Context) <-chan struct{} {
		// Ticker runs 1000 times.
		tickc := make(chan struct{})
		go func() {
			count := 0
			defer close(tickc)
			for {
				select {
				case tickc <- struct{}{}:
					count++
					if count == 1000 {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		return tickc
	}

	// Tests.
	count := 0
	err := wait.Poll(
		context.Background(),
		ticker,
		func() (bool, error) {
			count++
			if count == 500 {
				return true, nil
			}
			return false, nil
		},
	)
	verify.NoError(t, err)
	verify.Equal(t, count, 500)

	count = 0
	err = wait.Poll(
		context.Background(),
		ticker,
		func() (bool, error) {
			count++
			return false, nil
		},
	)
	verify.ErrorContains(t, err, "exceeded")
	verify.Equal(t, count, 1000, "exceeded with a count")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	err = wait.Poll(
		ctx,
		ticker,
		func() (bool, error) {
			time.Sleep(2 * time.Millisecond)
			return false, nil
		},
	)
	verify.ErrorContains(t, err, "cancelled")
}

// TestPanic tests the handling of panics during condition checks.
func TestPanic(t *testing.T) {
	count := 0
	err := wait.WithInterval(context.Background(), 10*time.Millisecond, func() (bool, error) {
		count++
		if count == 5 {
			panic("ouch at five o'clock")
		}
		return false, nil
	})
	verify.ErrorContains(t, err, "panic")
	verify.Equal(t, count, 5)
}

//--------------------
// HELPER
//--------------------

// mkChgTicker creates a ticker with a changing interval.
func mkChgTicker() wait.TickerFunc {
	interval := 5 * time.Millisecond
	return wait.MakeGenericIntervalTicker(func(in time.Duration) (out time.Duration, ok bool) {
		if in == 0 {
			return interval, true
		}
		out = in * 2
		if out > time.Second {
			return 0, false
		}
		return out, true
	})
}

// EOF
