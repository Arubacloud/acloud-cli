package cmd

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIsOperationalState(t *testing.T) {
	for _, s := range []string{StateActive, StateRunning, StateUsed, StateNotUsed} {
		if !isOperationalState(s) {
			t.Errorf("isOperationalState(%q) = false, want true", s)
		}
	}
	for _, s := range []string{StateInCreation, StateCreating, StateUpdating, StateError, StateFailed} {
		if isOperationalState(s) {
			t.Errorf("isOperationalState(%q) = true, want false", s)
		}
	}
}

func TestIsFailureState(t *testing.T) {
	for _, s := range []string{StateError, StateFailed, StateDeleted} {
		if !isFailureState(s) {
			t.Errorf("isFailureState(%q) = false, want true", s)
		}
	}
	for _, s := range []string{StateActive, StateInCreation, StateRunning} {
		if isFailureState(s) {
			t.Errorf("isFailureState(%q) = true, want false", s)
		}
	}
}

func TestWaitUntilActive_AlreadyActive(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateActive, nil }
	err := waitUntilActiveWithInterval(context.Background(), getter, "Cloud server", "my-vm", time.Millisecond)
	if err != nil {
		t.Errorf("expected nil for already-Active resource, got: %v", err)
	}
}

func TestWaitUntilActive_AlreadyRunning(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateRunning, nil }
	err := waitUntilActiveWithInterval(context.Background(), getter, "KaaS", "my-cluster", time.Millisecond)
	if err != nil {
		t.Errorf("expected nil for Running resource, got: %v", err)
	}
}

func TestWaitUntilActive_TransitoryThenActive(t *testing.T) {
	calls := 0
	getter := func(_ context.Context) (string, error) {
		calls++
		if calls < 3 {
			return StateInCreation, nil
		}
		return StateActive, nil
	}
	err := waitUntilActiveWithInterval(context.Background(), getter, "VPC", "my-vpc", time.Millisecond)
	if err != nil {
		t.Errorf("expected nil after InCreation→Active transition, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 getter calls, got %d", calls)
	}
}

func TestWaitUntilActive_FailureState(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateFailed, nil }
	err := waitUntilActiveWithInterval(context.Background(), getter, "DBaaS", "my-db", time.Millisecond)
	if err == nil {
		t.Fatal("expected error for Failed state, got nil")
	}
	if !contains(err.Error(), "failure state") {
		t.Errorf("expected 'failure state' in error, got: %v", err)
	}
}

func TestWaitUntilActive_ErrorState(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateError, nil }
	err := waitUntilActiveWithInterval(context.Background(), getter, "KMS", "my-kms", time.Millisecond)
	if err == nil {
		t.Fatal("expected error for Error state, got nil")
	}
	if !contains(err.Error(), "failure state") {
		t.Errorf("expected 'failure state' in error, got: %v", err)
	}
}

func TestWaitUntilActive_GetterError(t *testing.T) {
	getter := func(_ context.Context) (string, error) {
		return "", errors.New("API unavailable")
	}
	err := waitUntilActiveWithInterval(context.Background(), getter, "VPC", "my-vpc", time.Millisecond)
	if err == nil {
		t.Fatal("expected error when getter returns error, got nil")
	}
	if !contains(err.Error(), "polling") {
		t.Errorf("expected 'polling' in error, got: %v", err)
	}
}

func TestWaitUntilActive_ContextTimeout(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateInCreation, nil }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := waitUntilActiveWithInterval(ctx, getter, "Cloud server", "my-vm", time.Millisecond)
	if err == nil {
		t.Fatal("expected error on context timeout, got nil")
	}
	if !contains(err.Error(), "timed out") {
		t.Errorf("expected 'timed out' in error, got: %v", err)
	}
}

func TestWaitUntilActive_DeletedIsFailure(t *testing.T) {
	getter := func(_ context.Context) (string, error) { return StateDeleted, nil }
	err := waitUntilActiveWithInterval(context.Background(), getter, "Subnet", "my-subnet", time.Millisecond)
	if err == nil {
		t.Fatal("expected error for Deleted state, got nil")
	}
}
