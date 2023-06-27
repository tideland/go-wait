// Tideland Go Wait
//
// Copyright (C) 2019-2023 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package wait // import "tideland.dev/go/wait"

//--------------------
// IMPORTS
//--------------------

import (
	"context"
	"fmt"
	"math"

	"golang.org/x/time/rate"
)

//--------------------
// THROTTLE
//--------------------

// Task defines the signarure of a task to be processed.
type Task func() error

// Limit defines the rate limit of a throttle.
type Limit = rate.Limit

const (
	// Inf is the infinite rate limit.
	InfLimit = Limit(math.MaxFloat64)
)

// A Throttle limits the processing of tasks per second. It is configured with a
// limit and a burst. The limit is the maximum number of tasks per second and the
// burst the maximum number of tasks that can be processed at once. If the limit
// is InfLimit the throttle is not limited, if it is 0 no tasks can be processed.
type Throttle struct {
	limiter *rate.Limiter
}

// NewThrottle creates a new Throttle with the specified limit and burst.
func NewThrottle(limit Limit, burst int) *Throttle {
	return &Throttle{
		limiter: rate.NewLimiter(limit, burst),
	}
}

// Process processes a task under the context, waiting if necessary.
func (t *Throttle) Process(ctx context.Context, task Task) error {
	// Wait for the limiter to allow us to proceed.
	if err := t.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("wait for throttle limiter: %w", err)
	}
	// Process the task and return its error.
	return task()
}

// EOF
