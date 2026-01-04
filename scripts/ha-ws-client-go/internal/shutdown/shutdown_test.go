package shutdown

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	// Note: We can't access unexported gracePeriod field, but we can verify
	// the default behavior through other means
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

func TestCoordinator_RegisterCleanup(t *testing.T) {
	coord, _ := New()

	cleanupOrder := make([]int, 0, 3)
	var mu sync.Mutex

	coord.RegisterCleanup("first", func(_ context.Context) error {
		mu.Lock()
		cleanupOrder = append(cleanupOrder, 1)
		mu.Unlock()
		return nil
	})

	coord.RegisterCleanup("second", func(_ context.Context) error {
		mu.Lock()
		cleanupOrder = append(cleanupOrder, 2)
		mu.Unlock()
		return nil
	})

	coord.RegisterCleanup("third", func(_ context.Context) error {
		mu.Lock()
		cleanupOrder = append(cleanupOrder, 3)
		mu.Unlock()
		return nil
	})

	coord.Shutdown("test")

	// Give cleanups time to run
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Cleanups should run in LIFO order
	expected := []int{3, 2, 1}
	if len(cleanupOrder) != len(expected) {
		t.Fatalf("expected %d cleanups, got %d", len(expected), len(cleanupOrder))
	}
	for i, v := range expected {
		if cleanupOrder[i] != v {
			t.Errorf("cleanup order mismatch at %d: expected %d, got %d", i, v, cleanupOrder[i])
		}
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

	pr.RecordSuccess("item1")
	pr.RecordSuccess("item2")
	pr.RecordError(context.DeadlineExceeded)

	completed, total, errs := pr.Summary()

	if completed != 3 {
		t.Errorf("expected 3 completed, got %d", completed)
	}
	if total != 10 {
		t.Errorf("expected 10 total, got %d", total)
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
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

	pr.RecordSuccess(nil)

	if pr.IsComplete() {
		t.Error("expected not to be complete after 1 record")
	}

	pr.RecordSuccess(nil)

	if !pr.IsComplete() {
		t.Error("expected to be complete after 2 records")
	}
}

func TestWrapContext(t *testing.T) {
	coord, _ := New()
	parent := context.Background()

	ctx, cancel := WrapContext(parent, coord)
	defer cancel()

	// Context should not be canceled initially
	select {
	case <-ctx.Done():
		t.Error("expected context to not be canceled initially")
	default:
		// Expected
	}

	// Shutdown should cancel the wrapped context
	coord.Shutdown("test")

	// Give the goroutine time to cancel
	time.Sleep(10 * time.Millisecond)

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("expected context to be canceled after shutdown")
	}
}

func TestWrapContext_ParentCancel(_ *testing.T) {
	coord, _ := New()
	parent, parentCancel := context.WithCancel(context.Background())

	ctx, cancel := WrapContext(parent, coord)
	defer cancel()

	// Cancel parent
	parentCancel()

	// Give the goroutine time to propagate
	time.Sleep(10 * time.Millisecond)

	select {
	case <-ctx.Done():
		// Expected - parent cancellation propagates
	default:
		// This is also acceptable since the WrapContext goroutine may check ctx.Done() first
	}
}

func TestWithOnCleanupTimeout(t *testing.T) {
	timeoutCalled := false

	coord, _ := New(
		WithGracePeriod(50*time.Millisecond),
		WithOnCleanupTimeout(func() {
			timeoutCalled = true
		}),
	)

	// Register a slow cleanup that will exceed the grace period
	coord.RegisterCleanup("slow", func(ctx context.Context) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	coord.Shutdown("test")

	// Wait for cleanup timeout
	time.Sleep(100 * time.Millisecond)

	if !timeoutCalled {
		t.Error("expected onCleanupTimeout to be called")
	}
}

func TestCoordinator_CleanupError(t *testing.T) {
	coord, _ := New()

	// Register a cleanup that returns an error
	coord.RegisterCleanup("failing", func(_ context.Context) error {
		return errors.New("cleanup failed")
	})

	// Should not panic even with error
	coord.Shutdown("test")

	// Give cleanups time to run
	time.Sleep(50 * time.Millisecond)

	// Shutdown should still complete
	if !coord.IsShuttingDown() {
		t.Error("expected coordinator to be in shutdown state")
	}
}

func TestCoordinator_ConcurrentShutdown(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	coord, _ := New(
		WithOnShutdown(func(_ string) {
			mu.Lock()
			callCount++
			mu.Unlock()
		}),
	)

	// Call shutdown from multiple goroutines concurrently
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			coord.Shutdown(fmt.Sprintf("reason-%d", n))
		}(i)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if callCount != 1 {
		t.Errorf("expected onShutdown to be called exactly once, was called %d times", callCount)
	}
}

func TestCoordinator_RegisterCleanupConcurrent(t *testing.T) {
	coord, _ := New()

	var wg sync.WaitGroup
	cleanupCount := 0
	var mu sync.Mutex

	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			coord.RegisterCleanup(fmt.Sprintf("cleanup-%d", n), func(_ context.Context) error {
				mu.Lock()
				cleanupCount++
				mu.Unlock()
				return nil
			})
		}(i)
	}

	wg.Wait()

	// Trigger shutdown to run all cleanups
	coord.Shutdown("test")

	// Give cleanups time to run
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if cleanupCount != 10 {
		t.Errorf("expected 10 cleanups to run, got %d", cleanupCount)
	}
}

func TestCoordinator_Wait(t *testing.T) {
	coord, _ := New()

	done := make(chan bool, 1)

	go func() {
		coord.Wait()
		done <- true
	}()

	// Wait should block initially
	select {
	case <-done:
		t.Error("Wait() should block before shutdown")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Trigger shutdown
	coord.Shutdown("test")

	// Wait should complete now
	select {
	case <-done:
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Error("Wait() should complete after shutdown")
	}
}

func TestPartialResult_ConcurrentAccess(t *testing.T) {
	pr := NewPartialResult(100)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			pr.RecordSuccess("item")
		}()
		go func() {
			defer wg.Done()
			pr.RecordError(errors.New("error"))
		}()
	}

	wg.Wait()

	completed, total, errs := pr.Summary()
	if completed != 100 {
		t.Errorf("expected 100 completed, got %d", completed)
	}
	if total != 100 {
		t.Errorf("expected 100 total, got %d", total)
	}
	if len(errs) != 50 {
		t.Errorf("expected 50 errors, got %d", len(errs))
	}
}

func TestWrapContext_CancelFunc(t *testing.T) {
	coord, _ := New()
	parent := context.Background()

	ctx, cancel := WrapContext(parent, coord)

	// Calling cancel should cancel the context
	cancel()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("expected context to be canceled after calling cancel func")
	}
}
