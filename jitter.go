// Tideland Go Wait
//
// Copyright (C) 2019-2025 Frank Mueller / Tideland / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package wait


import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"time"
)

// MakeJitteringTicker returns a ticker signalling in jittering intervals. This
// avoids converging on periadoc behavior during condition check. The returned
// interval jitters inside the given interval and starts with the given offset.
// The ticker stops after reaching the timeout.
func MakeJitteringTicker(interval, offset, timeout time.Duration) TickerFunc {
	start := time.Now()
	deadline := start.Add(timeout)

	// Sanitize
	if interval < time.Millisecond {
		interval = time.Millisecond
	}
	if offset < 0 {
		offset = 0
	}
	if interval > time.Duration(math.MaxInt64)-offset {
		interval = time.Duration(math.MaxInt64) - offset
	}

	next := start

	changer := func(_ time.Duration) (time.Duration, bool) {
		now := time.Now()
		if now.After(deadline) {
			return 0, false
		}

		// Generate jitter in range [0, interval)
		jitterRange := interval
		if deadline.Sub(now) < offset+jitterRange {
			jitterRange = deadline.Sub(now) - offset
			if jitterRange < 1 {
				return 0, false
			}
		}

		bigInt, err := rand.Int(rand.Reader, big.NewInt(jitterRange.Nanoseconds()))
		if err != nil {
			return 0, false
		}

		jitter := time.Duration(bigInt.Int64())
		wait := offset + jitter

		next = next.Add(wait)
		delay := max(time.Until(next), 0)
		return delay, true
	}

	return MakeGenericIntervalTicker(changer)
}

// WithJitter is convenience for Poll() with MakeJitteringTicker().
func WithJitter(
	ctx context.Context,
	interval, offset, timeout time.Duration,
	condition ConditionFunc,
) error {
	return Poll(
		ctx,
		MakeJitteringTicker(interval, offset, timeout),
		condition,
	)
}

