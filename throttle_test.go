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
	"errors"
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
		name    string
		events  int
		perCall int
		limit   wait.Limit
		burst   int
		fail    bool
		cancel  bool
		timeout time.Duration
		err     string
	}{
		{
			name:    "single-burst-single-event",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   1,
		},
		{
			name:    "larger-burst-single-event",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   5,
		},
		{
			name:    "single-burst-multiple-events",
			events:  55,
			perCall: 5,
			limit:   20.0,
			burst:   1,
			err:     "event(s) exceeds throttle burst size 1",
		},
		{
			name:    "cancel-before-processing",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   1,
			cancel:  true,
			err:     "event(s) throttle context already done",
		},
		{
			name:    "exceeding-context-timeout",
			events:  55,
			perCall: 1,
			limit:   20.0,
			burst:   1,
			timeout: time.Millisecond,
			err:     "event(s) would exceed throttle context deadline",
		},
		{
			name:    "exceeding-event-timeout",
			events:  10,
			perCall: 1,
			limit:   1,
			burst:   1,
			cancel:  true,
			timeout: time.Millisecond,
			err:     "event(s) throttle context timed out or cancelled",
		},
		{
			name:    "error-in-event",
			events:  5,
			perCall: 1,
			limit:   5,
			burst:   1,
			fail:    true,
			err:     "processing event 0 returned error: ouch",
		},
	}
	// Run tests.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.SetFailable(t)
			// Preparings.
			var rushesDone int64
			var errorsReturned int64
			rushes := test.events / test.perCall
			event := func() error {
				if test.fail {
					return errors.New("ouch")
				}
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
					var cancel func()
					switch {
					case test.cancel && test.timeout == 0:
						// Cancel context before it is used.
						ctx, cancel = context.WithCancel(ctx)
						cancel()
						cancel = nil
					case !test.cancel && test.timeout > 0:
						// Set context with a timeout.
						ctx, cancel = context.WithTimeout(ctx, test.timeout)
						defer cancel()
					case test.cancel && test.timeout > 0:
						// Cancel the context while event is waiting.
						ctx, cancel = context.WithCancel(ctx)
						go func() {
							time.Sleep(test.timeout)
							cancel()
						}()
					}
					err := throttle.Process(ctx, events...)
					if test.err != "" && err != nil {
						atomic.AddInt64(&errorsReturned, 1)
						// assert.Logf("processing error: %v", err.Error())
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
