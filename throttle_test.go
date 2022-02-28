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
	"fmt"
	"sync/atomic"
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
		name     string
		events   int
		perCall  int
		limit    wait.Limit
		burst    int
		cancel   bool
		deadline time.Duration
		err      string
	}{
		{
			name:    "single burst and single event",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   1,
		},
		{
			name:    "larger burst and single event",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   5,
		},
		{
			name:    "single burst and multiple events",
			events:  55,
			perCall: 5,
			limit:   20.0,
			burst:   1,
			err:     "event(s) exceeds throttle burst size 1",
		},
		{
			name:    "cancel before processing",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   1,
			cancel:  true,
			err:     "event(s) throttle context already done",
		},
		{
			name:     "exceeding deadline",
			events:   55,
			perCall:  1,
			limit:    20.0,
			burst:    1,
			deadline: time.Millisecond,
			err:      "event(s) would exceed throttle context deadline",
		},
	}
	// Run tests.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.SetFailable(t)
			assert.Logf("running <<%s>>", test.name)
			// Preparings.
			var rushesDone int64
			var errorsReturned int64
			rushes := test.events / test.perCall
			event := func() error {
				return nil
			}
			throttle := wait.NewThrottle(test.limit, test.burst)
			rushing := func() {
				rush := func() {
					ctx := context.Background()
					events := []wait.Event{}
					for j := 0; j < test.perCall; j++ {
						events = append(events, event)
					}
					if test.cancel {
						// Cancel context before used.
						var cancel func()
						ctx, cancel = context.WithCancel(ctx)
						cancel()
					}
					if test.deadline > 0 {
						// Set context deadline.
						var cancel func()
						ctx, cancel = context.WithTimeout(ctx, test.deadline)
						defer cancel()
					}
					err := throttle.Process(ctx, events...)
					if test.err != "" {
						atomic.AddInt64(&errorsReturned, 1)
						assert.ErrorContains(err, test.err)
					}
					atomic.AddInt64(&rushesDone, 1)
				}
				for i := 0; i < rushes; i++ {
					go rush()
				}
			}
			// Start rushing.
			start := time.Now()

			go rushing()

			ticker := time.NewTicker(5 * time.Millisecond)
			defer ticker.Stop()

			for {
				<-ticker.C

				l := throttle.Limit()
				b := throttle.Burst()

				assert.Equal(l, wait.Limit(test.limit))
				assert.Equal(b, test.burst)

				if atomic.LoadInt64(&rushesDone) == int64(rushes) {
					break
				}
			}

			duration := time.Now().Sub(start)
			seconds := float64(test.events) / float64(test.limit)
			info := fmt.Sprintf("duration is %.4f, not %.4fs (+/- 0.25s)", duration.Seconds(), seconds)

			if er := atomic.LoadInt64(&errorsReturned); er > 0 {
				// In case of errors look at their count.
				assert.True(er <= int64(rushes))
			} else {
				// Otherwise look at the count of the rushes and the time.
				assert.Equal(atomic.LoadInt64(&rushesDone), int64(rushes))
				assert.About(duration.Seconds(), seconds, 0.25, info)
			}
		})
	}
}

// EOF
