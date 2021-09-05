// Tideland Go Wait - Unit Tests
//
// Copyright (C) 2019-2021 Frank Mueller / Tideland / Oldenburg / Germany
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

	"tideland.dev/go/audit/asserts"

	"tideland.dev/go/wait"
)

//--------------------
// TESTS
//--------------------

// TestPolls verifies Poll() with different parameters.
func TestChangingInterval(t *testing.T) {
	assert := asserts.NewTesting(t, asserts.FailStop)
	tests := []struct {
		name     string
		ticker   func() wait.TickerFunc
		duration time.Duration
		count    int
		err      string
	}{
		{
			name: "interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeIntervalTicker(5 * time.Millisecond)
			},
			count: 5,
		}, {
			name: "interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeIntervalTicker(5 * time.Millisecond)
			},
			duration: 50 * time.Millisecond,
			err:      "context has been cancelled with error",
		}, {
			name:   "changing-interval-poll-success",
			ticker: mkChgTicker,
			count:  5,
		}, {
			name:   "changing-interval-poll-ticker-exceeds",
			ticker: mkChgTicker,
			err:    "ticker exceeded while waiting for the condition",
		}, {
			name:     "changing-interval-poll-context-cancelled",
			ticker:   mkChgTicker,
			duration: 50 * time.Millisecond,
			err:      "context has been cancelled with error",
		}, {
			name: "max-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			count: 5,
		}, {
			name: "max-interval-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			err: "ticker exceeded while waiting for the condition",
		}, {
			name: "max-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeMaxIntervalsTicker(5*time.Millisecond, 10)
			},
			duration: 50 * time.Millisecond,
			err:      "context has been cancelled with error",
		}, {
			name: "deadlined-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			count: 5,
		}, {
			name: "deadlined-interval-poll-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			err: "ticker exceeded while waiting for the condition",
		}, {
			name: "deadline-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeDeadlinedIntervalTicker(5*time.Millisecond, time.Now().Add(55*time.Millisecond))
			},
			duration: 50 * time.Millisecond,
			err:      "context has been cancelled with error",
		}, {
			name: "expiring-interval-poll-success",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			count: 5,
		}, {
			name: "expiring-interval-poll-ticker-exceeds",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			err: "ticker exceeded while waiting for the condition",
		}, {
			name: "expiring-interval-poll-context-cancelled",
			ticker: func() wait.TickerFunc {
				return wait.MakeExpiringIntervalTicker(5*time.Millisecond, 55*time.Millisecond)
			},
			duration: 50 * time.Millisecond,
			err:      "context has been cancelled with error",
		},
	}
	// Run tests.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.SetFailable(t)
			ctx := context.Background()
			if test.duration != 0 {
				ctx, _ = context.WithTimeout(ctx, test.duration)
			}
			count := 0
			err := wait.Poll(
				ctx,
				test.ticker(),
				func() (bool, error) {
					count++
					if count == test.count {
						return true, nil
					}
					return false, nil
				},
			)
			if test.err == "" {
				assert.NoError(err)
				assert.Equal(count, test.count)
			} else {
				assert.ErrorContains(err, test.err)
			}
		})
	}
}

// TestWithInterval verifies WithInterval().
func TestWithInterval(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Test.
	count := 0
	err := wait.WithInterval(context.Background(), 20*time.Millisecond, func() (bool, error) {
		count++
		if count == 5 {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(err)
	assert.Equal(count, 5)
}

// TestPollWithMaxIntervals tests the polling of conditions in a maximum
// number of intervals.
func TestPollWithMaxInterval(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Tests.
	count := 0
	err := wait.WithMaxIntervals(context.Background(), 20*time.Millisecond, 10, func() (bool, error) {
		count++
		if count == 5 {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(err)
	assert.Equal(count, 5)
}

// TestPollWithTimeout tests the polling of conditions with timeouts.
func TestPollWithTimeout(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Tests.
	count := 0
	err := wait.WithTimeout(context.Background(), 5*time.Millisecond, 55*time.Millisecond, func() (bool, error) {
		count++
		if count == 5 {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(err)
	assert.Equal(count, 5)
}

// TestPollWithJitter tests the polling of conditions in a maximum
// number of intervals.
func TestPollWithJitter(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Tests.
	timestamps := []time.Time{}
	err := wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(50*time.Millisecond, 1.0, 1250*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			if len(timestamps) == 10 {
				return true, nil
			}
			return false, nil
		},
	)
	assert.NoError(err)
	assert.Length(timestamps, 10)
	for i := 1; i < 10; i++ {
		diff := timestamps[i].Sub(timestamps[i-1])
		// 10% upper tolerance.
		assert.Range(diff, 50*time.Millisecond, 110*time.Millisecond)
	}

	timestamps = []time.Time{}
	err = wait.WithJitter(context.Background(), 50*time.Millisecond, 1.0, 1250*time.Millisecond, func() (bool, error) {
		timestamps = append(timestamps, time.Now())
		if len(timestamps) == 10 {
			return true, nil
		}
		return false, nil
	})
	assert.NoError(err)
	assert.Length(timestamps, 10)
	for i := 1; i < 10; i++ {
		diff := timestamps[i].Sub(timestamps[i-1])
		// 10% upper tolerance.
		assert.Range(diff, 50*time.Millisecond, 110*time.Millisecond)
	}

	timestamps = []time.Time{}
	err = wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(50*time.Millisecond, 1.0, 1250*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	assert.ErrorContains(err, "exceeded")
	assert.Range(len(timestamps), 10, 25)

	timestamps = []time.Time{}
	err = wait.Poll(
		context.Background(),
		wait.MakeJitteringTicker(50*time.Millisecond, 1.0, -10*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	assert.ErrorContains(err, "exceeded")
	assert.Empty(timestamps)

	timestamps = []time.Time{}
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()
	err = wait.Poll(
		ctx,
		wait.MakeJitteringTicker(50*time.Millisecond, 1.0, 1250*time.Millisecond),
		func() (bool, error) {
			timestamps = append(timestamps, time.Now())
			return false, nil
		},
	)
	assert.ErrorContains(err, "cancelled")
}

// TestPoll tests the polling of conditions with a user-defined ticker.
func TestPoll(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)
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
	assert.NoError(err)
	assert.Equal(count, 500)

	count = 0
	err = wait.Poll(
		context.Background(),
		ticker,
		func() (bool, error) {
			count++
			return false, nil
		},
	)
	assert.ErrorContains(err, "exceeded")
	assert.Equal(count, 1000, "exceeded with a count")

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
	assert.ErrorContains(err, "cancelled")
}

// TestPanic tests the handling of panics during condition checks.
func TestPanic(t *testing.T) {
	// Init.
	assert := asserts.NewTesting(t, asserts.FailStop)

	// Test.
	count := 0
	err := wait.WithInterval(context.Background(), 10*time.Millisecond, func() (bool, error) {
		count++
		if count == 5 {
			panic("ouch at five o'clock")
		}
		return false, nil
	})
	assert.ErrorContains(err, "panic")
	assert.Equal(count, 5)
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
