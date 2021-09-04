// Tideland Go Wait
//
// Copyright (C) 2019-2021 Frank Mueller / Tideland / Oldenburg / Germany
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
	"time"
)

//--------------------
// POLL
//--------------------

// ConditionFunc has to be implemented for checking the wanted condition. A positive
// condition will return true and nil, a negative false and nil. In case of failure
// during the check false and the error have to be returned. The function will
// be used by the poll functions.
type ConditionFunc func() (bool, error)

// Poll checks the condition until it returns true or an error. The ticker
// sends signals whenever the condition shall be checked. It closes the returned
// channel when the polling shall stop.
func Poll(ctx context.Context, ticker TickerFunc, condition ConditionFunc) error {
	tickCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tickc := ticker(tickCtx)
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context has been cancelled with error: %v", ctx.Err())
		case _, open := <-tickc:
			// Ticker sent a signal to check for condition.
			if !open {
				// Oh, ticker tells to end.
				return fmt.Errorf("ticker exceeded while waiting for the condition")
			}
			ok, err := check(condition)
			if err != nil {
				// ConditionFunc has an error.
				return fmt.Errorf("poll condition returned error: %v", err)
			}
			if ok {
				// ConditionFunc is happy.
				return nil
			}
		}
	}
}

// WithInterval is convenience for Poll() with MakeIntervalTicker().
func WithInterval(
	ctx context.Context,
	interval time.Duration,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeIntervalTicker(interval),
		condition,
	)
}

// WithMaxIntervals is convenience for Poll() with MakeMaxIntervalsTicker().
func WithMaxIntervals(
	ctx context.Context,
	interval time.Duration,
	max int,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeMaxIntervalsTicker(interval, max),
		condition,
	)
}

// WithDeadline is convenience for Poll() with MakeDeadlinedIntervalTicker().
func WithDeadline(
	ctx context.Context,
	interval time.Duration,
	deadline time.Time,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeDeadlinedIntervalTicker(interval, deadline),
		condition,
	)
}

// WithTimeout is convenience for Poll() with MakeExpiringIntervalTicker().
func WithTimeout(
	ctx context.Context,
	interval, timeout time.Duration,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeExpiringIntervalTicker(interval, timeout),
		condition,
	)
}

// WithJitter is convenience for Poll() with MakeJitteringTicker().
func WithJitter(
	ctx context.Context,
	interval time.Duration,
	factor float64,
	timeout time.Duration,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeJitteringTicker(interval, factor, timeout),
		condition,
	)
}

//--------------------
// PRIVATE HELPER
//--------------------

// check runs the condition catching potential panics and returns
// them as failure.
func check(condition ConditionFunc) (ok bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			err = fmt.Errorf("panic during condition check: %v", r)
		}
	}()
	ok, err = condition()
	return
}

// EOF
