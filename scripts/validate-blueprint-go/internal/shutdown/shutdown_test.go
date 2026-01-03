package shutdown

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	coord, ctx := New()

	if coord == nil {
		t.Fatal("expected non-nil coordinator")
	}
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if coord.gracePeriod != DefaultGracePeriod {
		t.Errorf("expected grace period %v, got %v", DefaultGracePeriod, coord.gracePeriod)
	}
}

func TestNewWithOptions(t *testing.T) {
	gracePeriod := 10 * time.Second
	shutdownCalled := false

	coord, _ := New(
		WithGracePeriod(gracePeriod),
		WithOnShutdown(func(_ string) {
			shutdownCalled = true
		}),
	)

	if coord.gracePeriod != gracePeriod {
		t.Errorf("expected grace period %v, got %v", gracePeriod, coord.gracePeriod)
	}

	coord.Shutdown("test")

	if !shutdownCalled {
		t.Error("expected onShutdown to be called")
	}
}

func TestCoordinator_Shutdown(t *testing.T) {
	coord, ctx := New()

	if coord.IsShuttingDown() {
		t.Error("expected not to be shutting down initially")
	}

	coord.Shutdown("test reason")

	if !coord.IsShuttingDown() {
		t.Error("expected to be shutting down")
	}

	if coord.ShutdownReason() != "test reason" {
		t.Errorf("expected reason 'test reason', got %q", coord.ShutdownReason())
	}

	// Context should be canceled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("expected context to be canceled")
	}
}

func TestCoordinator_ShutdownOnce(t *testing.T) {
	callCount := 0
	coord, _ := New(
		WithOnShutdown(func(_ string) {
			callCount++
		}),
	)

	// Call shutdown multiple times
	coord.Shutdown("first")
	coord.Shutdown("second")
	coord.Shutdown("third")

	if callCount != 1 {
		t.Errorf("expected onShutdown to be called once, was called %d times", callCount)
	}

	if coord.ShutdownReason() != "first" {
		t.Errorf("expected first shutdown reason to be preserved, got %q", coord.ShutdownReason())
	}
}

func TestCoordinator_ShutdownChan(t *testing.T) {
	coord, _ := New()

	// Channel should not be closed initially
	select {
	case <-coord.ShutdownChan():
		t.Error("expected channel to not be closed initially")
	default:
		// Expected
	}

	coord.Shutdown("test")

	// Channel should be closed after shutdown
	select {
	case <-coord.ShutdownChan():
		// Expected
	default:
		t.Error("expected channel to be closed after shutdown")
	}
}

func TestCoordinator_Context(t *testing.T) {
	coord, ctx := New()

	if coord.Context() != ctx {
		t.Error("expected Context() to return the same context")
	}
}

func TestPartialResult(t *testing.T) {
	pr := NewPartialResult(10)

	pr.RecordPass("item1")
	pr.RecordPass("item2")
	pr.RecordFail("item3", "some error")

	completed, total, passed, failed, errors := pr.Summary()

	if completed != 3 {
		t.Errorf("expected 3 completed, got %d", completed)
	}
	if total != 10 {
		t.Errorf("expected 10 total, got %d", total)
	}
	if passed != 2 {
		t.Errorf("expected 2 passed, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}

	if pr.IsComplete() {
		t.Error("expected not to be complete")
	}
}

func TestPartialResult_IsComplete(t *testing.T) {
	pr := NewPartialResult(2)

	if pr.IsComplete() {
		t.Error("expected not to be complete initially")
	}

	pr.RecordPass("item1")

	if pr.IsComplete() {
		t.Error("expected not to be complete after 1 record")
	}

	pr.RecordPass("item2")

	if !pr.IsComplete() {
		t.Error("expected to be complete after 2 records")
	}
}
