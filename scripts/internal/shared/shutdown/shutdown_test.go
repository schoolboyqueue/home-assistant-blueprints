package shutdown

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewCoordinator(t *testing.T) {
	coord, ctx := New()

	if coord == nil {
		t.Fatal("New should return a coordinator")
	}
	if ctx == nil {
		t.Fatal("New should return a context")
	}
	if coord.gracePeriod != DefaultGracePeriod {
		t.Errorf("Default grace period = %v, want %v", coord.gracePeriod, DefaultGracePeriod)
	}
}

func TestWithGracePeriod(t *testing.T) {
	customPeriod := 10 * time.Second
	coord, _ := New(WithGracePeriod(customPeriod))

	if coord.gracePeriod != customPeriod {
		t.Errorf("Grace period = %v, want %v", coord.gracePeriod, customPeriod)
	}
}

func TestWithOnShutdown(t *testing.T) {
	called := false
	var receivedReason string
	coord, _ := New(WithOnShutdown(func(reason string) {
		called = true
		receivedReason = reason
	}))

	coord.Shutdown("test reason")

	if !called {
		t.Error("OnShutdown callback should be called")
	}
	if receivedReason != "test reason" {
		t.Errorf("Shutdown reason = %q, want %q", receivedReason, "test reason")
	}
}

func TestShutdownOnlyOnce(t *testing.T) {
	callCount := 0
	coord, _ := New(WithOnShutdown(func(string) {
		callCount++
	}))

	coord.Shutdown("first")
	coord.Shutdown("second")
	coord.Shutdown("third")

	if callCount != 1 {
		t.Errorf("OnShutdown called %d times, want 1", callCount)
	}
}

func TestShutdownCancelsContext(t *testing.T) {
	coord, ctx := New()

	select {
	case <-ctx.Done():
		t.Fatal("Context should not be done before shutdown")
	default:
	}

	coord.Shutdown("test")

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(time.Second):
		t.Fatal("Context should be canceled after shutdown")
	}
}

func TestShutdownChan(t *testing.T) {
	coord, _ := New()

	// Should not be closed initially
	select {
	case <-coord.ShutdownChan():
		t.Fatal("ShutdownChan should not be closed before shutdown")
	default:
	}

	coord.Shutdown("test")

	// Should be closed after shutdown
	select {
	case <-coord.ShutdownChan():
		// Expected
	case <-time.After(time.Second):
		t.Fatal("ShutdownChan should be closed after shutdown")
	}
}

func TestIsShuttingDown(t *testing.T) {
	coord, _ := New()

	if coord.IsShuttingDown() {
		t.Error("IsShuttingDown should be false before shutdown")
	}

	coord.Shutdown("test")

	if !coord.IsShuttingDown() {
		t.Error("IsShuttingDown should be true after shutdown")
	}
}

func TestShutdownReason(t *testing.T) {
	coord, _ := New()

	if coord.ShutdownReason() != "" {
		t.Error("ShutdownReason should be empty before shutdown")
	}

	coord.Shutdown("test reason")

	if coord.ShutdownReason() != "test reason" {
		t.Errorf("ShutdownReason = %q, want %q", coord.ShutdownReason(), "test reason")
	}
}

func TestRegisterCleanup(t *testing.T) {
	coord, _ := New(WithGracePeriod(time.Second))

	var order []int
	var mu sync.Mutex

	coord.RegisterCleanup("first", func(context.Context) error {
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		return nil
	})
	coord.RegisterCleanup("second", func(context.Context) error {
		mu.Lock()
		order = append(order, 2)
		mu.Unlock()
		return nil
	})
	coord.RegisterCleanup("third", func(context.Context) error {
		mu.Lock()
		order = append(order, 3)
		mu.Unlock()
		return nil
	})

	coord.Shutdown("test")

	// Give time for cleanups to complete
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Should be LIFO order
	expected := []int{3, 2, 1}
	if len(order) != len(expected) {
		t.Fatalf("Cleanup count = %d, want %d", len(order), len(expected))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("Cleanup order[%d] = %d, want %d", i, order[i], v)
		}
	}
}

func TestCleanupTimeout(t *testing.T) {
	timeoutCalled := false
	coord, _ := New(
		WithGracePeriod(100*time.Millisecond),
		WithOnCleanupTimeout(func() {
			timeoutCalled = true
		}),
	)

	coord.RegisterCleanup("slow", func(ctx context.Context) error {
		select {
		case <-time.After(time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	coord.Shutdown("test")

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	if !timeoutCalled {
		t.Error("OnCleanupTimeout should be called")
	}
}

func TestWrapContext(t *testing.T) {
	coord, _ := New()
	parent := context.Background()

	ctx, cancel := WrapContext(parent, coord)
	defer cancel()

	// Context should not be done
	select {
	case <-ctx.Done():
		t.Fatal("Wrapped context should not be done before shutdown")
	default:
	}

	// Shutdown should cancel the wrapped context
	coord.Shutdown("test")

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(time.Second):
		t.Fatal("Wrapped context should be canceled after shutdown")
	}
}

func TestWrapContextParentCancel(t *testing.T) {
	coord, _ := New()
	parent, parentCancel := context.WithCancel(context.Background())

	ctx, cancel := WrapContext(parent, coord)
	defer cancel()

	// Cancel the parent
	parentCancel()

	select {
	case <-ctx.Done():
		// Expected - parent cancellation should propagate
	case <-time.After(time.Second):
		t.Fatal("Wrapped context should be canceled when parent is canceled")
	}
}

func TestPartialResult(t *testing.T) {
	p := NewPartialResult(5)

	if p.IsComplete() {
		t.Error("New PartialResult should not be complete")
	}

	// Record some successes
	p.RecordSuccess("result1")
	p.RecordSuccess("result2")

	// Record some errors
	p.RecordError(fmt.Errorf("error1"))

	completed, total, errors := p.Summary()
	if completed != 3 {
		t.Errorf("Completed = %d, want 3", completed)
	}
	if total != 5 {
		t.Errorf("Total = %d, want 5", total)
	}
	if len(errors) != 1 {
		t.Errorf("Error count = %d, want 1", len(errors))
	}

	if p.LastResult() != "result2" {
		t.Errorf("LastResult = %v, want 'result2'", p.LastResult())
	}
}

func TestPartialResultWithCounts(t *testing.T) {
	p := NewPartialResult(4)

	p.RecordPass("item1")
	p.RecordPass("item2")
	p.RecordFail("item3", "invalid")
	p.RecordFail("item4", "missing")

	completed, total, passed, failed, messages := p.SummaryWithCounts()
	if completed != 4 {
		t.Errorf("Completed = %d, want 4", completed)
	}
	if total != 4 {
		t.Errorf("Total = %d, want 4", total)
	}
	if passed != 2 {
		t.Errorf("Passed = %d, want 2", passed)
	}
	if failed != 2 {
		t.Errorf("Failed = %d, want 2", failed)
	}
	if len(messages) != 2 {
		t.Errorf("Messages count = %d, want 2", len(messages))
	}

	if !p.IsComplete() {
		t.Error("Should be complete")
	}
}

func TestPartialResultConcurrency(t *testing.T) {
	p := NewPartialResult(100)
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				p.RecordSuccess(i)
			} else {
				p.RecordError(fmt.Errorf("error %d", i))
			}
		}(i)
	}

	wg.Wait()

	completed, total, _ := p.Summary()
	if completed != 100 {
		t.Errorf("Completed = %d, want 100", completed)
	}
	if total != 100 {
		t.Errorf("Total = %d, want 100", total)
	}
}

func TestCoordinatorConcurrency(t *testing.T) {
	coord, ctx := New()

	var wg sync.WaitGroup
	var shutdownCount int32

	// Multiple goroutines trying to shutdown
	for i := range 10 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			coord.Shutdown(fmt.Sprintf("reason %d", i))
			atomic.AddInt32(&shutdownCount, 1)
		}(i)
	}

	// Goroutines checking shutdown state
	for range 10 {
		wg.Go(func() {
			_ = coord.IsShuttingDown()
			_ = coord.ShutdownReason()
		})
	}

	wg.Wait()

	// Context should be canceled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be done after shutdown")
	}

	// All goroutines should have completed
	if atomic.LoadInt32(&shutdownCount) != 10 {
		t.Error("All shutdown calls should complete")
	}
}

func TestContext(t *testing.T) {
	coord, ctx := New()

	if coord.Context() != ctx {
		t.Error("Context() should return the context from New()")
	}
}
