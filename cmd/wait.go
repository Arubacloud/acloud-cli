package cmd

import (
	"context"
	"fmt"
	"os"
	"time"
)

// StateGetter retrieves the current state string of a resource.
// Implementations call the SDK Get method for a specific resource type.
type StateGetter func(ctx context.Context) (string, error)

// isOperationalState returns true for settled non-failure states that mean the
// resource is ready. Mirrors types.State.IsOperational/IsAvailable in the SDK.
func isOperationalState(s string) bool {
	switch s {
	case StateActive, StateRunning, StateUsed, StateNotUsed:
		return true
	}
	return false
}

// isFailureState returns true for terminal failure states.
// Mirrors types.State.IsFailure from the SDK.
func isFailureState(s string) bool {
	switch s {
	case StateError, StateFailed, StateDeleted:
		return true
	}
	return false
}

// defaultWaitInterval is the poll interval used by WaitUntilActive.
// Tests override it via waitUntilActiveWithInterval.
const defaultWaitInterval = 5 * time.Second

// WaitUntilActive polls the resource via getter every 5 s until it reaches an
// operational state or a failure state, or the ctx deadline fires.
//
// Progress is printed to stderr so it does not pollute stdout (which may be
// parsed by callers using -o json). Use --timeout to control the deadline.
func WaitUntilActive(ctx context.Context, getter StateGetter, kind, name string) error {
	return waitUntilActiveWithInterval(ctx, getter, kind, name, defaultWaitInterval)
}

// waitUntilActiveWithInterval is the testable core, accepting a custom poll
// interval. Pass a small duration in tests to avoid sleeping.
func waitUntilActiveWithInterval(ctx context.Context, getter StateGetter, kind, name string, interval time.Duration) error {
	start := time.Now()

	for {
		state, err := getter(ctx)
		if err != nil {
			return fmt.Errorf("polling %s '%s': %w", kind, name, err)
		}

		elapsed := time.Since(start).Truncate(time.Second)

		if isOperationalState(state) {
			fmt.Fprintf(os.Stderr, "\r%s '%s' is %s after %s.\n", kind, name, state, elapsed)
			return nil
		}
		if isFailureState(state) {
			return fmt.Errorf("%s '%s' reached failure state %q after %s", kind, name, state, elapsed)
		}

		fmt.Fprintf(os.Stderr, "\rWaiting for %s '%s' to become Active... [%s]  ", kind, name, elapsed)

		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for %s '%s' to become Active after %s", kind, name, elapsed)
		case <-time.After(interval):
		}
	}
}
