// Tideland Go Wait
//
// Copyright (C) 2019-2022 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package wait // import "tideland.dev/go/wait"

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"math"
	"sync"
	"time"
)

//--------------------
// CONSTANTS
//--------------------

const (
	InfinityLimit    = math.MaxInt64
	InfinityDuration = time.Duration(math.MaxInt64)

	eventLoad = 1
)

//--------------------
// LIMIT
//--------------------

// Limit defines the number of allowed job starts per second by the throttle.
type Limit int64

//--------------------
// THROTTLE
//--------------------

// Event wraps the event to be processed inside a function executed by a throttle.
type Event func() error

// Throttle controls the maximum number of processed events per second.
type Throttle struct {
	mu            sync.RWMutex
	limit         Limit
	retryInterval time.Duration
	factor        float64
	limitedLoad   int64
	currentLoad   int64
	eventDuration time.Duration
}

// NewThrottle creates a new throttle allowing a limited event processing per second.
func NewThrottle(limit Limit) *Throttle {
	if limit < 1 {
		limit = 1
	}
	eventDuration := time.Second / time.Duration(limit)
	return &Throttle{
		limit:         limit,
		retryInterval: eventDuration,
		factor:        0.1,
		limitedLoad:   int64(limit),
		eventDuration: eventDuration,
	}
}

// Process processes an event if the throttle has capacity.
func (t *Throttle) Process(ctx context.Context, event Event) error {
	// Check if the context is already cancelled.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Calculate a possible timeout.
	now := time.Now()
	timeout := InfinityDuration
	if deadline, ok := ctx.Deadline(); ok {
		timeout = deadline.Sub(now)
	}
	// Wait until allowed.
	if err := WithJitter(ctx, t.retryInterval, 0.1, timeout, t.isAllowed); err != nil {
		return err
	}
	return t.process(event)
}

// LimitedLoad allows to retrieve the limited usage of the throttle in events per interval.
func (t *Throttle) LimitedLoad() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.limitedLoad
}

// CurrentLoad allows to retrieve the current usage of the throttle in events per interval.
func (t *Throttle) CurrentLoad() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLoad
}

// isAllowed checks if the throttle allows the event processing. It implements the
// ConditionFunc for waiting functions.
func (t *Throttle) isAllowed() (bool, error) {
	t.mu.RLock()
	limitedLoad := t.limitedLoad
	currentLoad := t.currentLoad
	t.mu.RUnlock()
	return currentLoad < limitedLoad, nil
}

// process executes the event and adjusts the current load during execution.
func (t *Throttle) process(event Event) error {
	// Before processing.
	t.mu.Lock()
	t.currentLoad += eventLoad
	eventDuration := t.eventDuration
	t.mu.Unlock()

	// Processing.
	before := time.Now()
	err := event()
	after := time.Now()
	duration := after.Sub(before)

	if duration < eventDuration {
		// Slow down a bit.
		slowDown := time.Duration(eventDuration.Nanoseconds() - duration.Nanoseconds())
		time.Sleep(slowDown)
	}

	// After processing.
	t.mu.Lock()
	t.currentLoad -= eventLoad
	t.mu.Unlock()

	return err
}

// EOF
