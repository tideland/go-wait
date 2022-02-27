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
		name   string
		runs   int
		limit  wait.Limit
		burst  int
		events int
		err    string
	}{
		{
			name:   "single burst loop",
			runs:   110,
			limit:  20.0,
			burst:  1,
			events: 1,
		},
	}
	// Run tests.
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.SetFailable(t)
			// Preparings.
			var rushes int64
			event := func() error {
				atomic.AddInt64(&rushes, 1)
				return nil
			}
			throttle := wait.NewThrottle(test.limit, test.burst)
			rushing := func() {
				rush := func() {
					ctx := context.Background()
					events := []wait.Event{}
					for j := 0; j < test.events; j++ {
						events = append(events, event)
					}
					throttle.Process(ctx, events...)
				}
				for i := 0; i < test.runs/test.events; i++ {
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

				rs := atomic.LoadInt64(&rushes)
				l := throttle.Limit()
				b := throttle.Burst()

				assert.Equal(l, wait.Limit(test.limit))
				assert.Equal(b, test.burst)

				if rs >= int64(test.runs) {
					break
				}
			}

			duration := time.Now().Sub(start)
			seconds := float64(test.runs) / float64(test.limit)
			info := fmt.Sprintf("duration is %.4f, not %.4fs (+/- 0.25s)", duration.Seconds(), seconds)

			assert.Equal(rushes, int64(test.runs))
			assert.About(duration.Seconds(), seconds, 0.25, info)
		})
	}
}

// EOF
