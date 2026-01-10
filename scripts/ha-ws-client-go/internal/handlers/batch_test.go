package handlers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchError(t *testing.T) {
	t.Parallel()

	t.Run("formats error correctly", func(t *testing.T) {
		t.Parallel()
		err := &BatchError{
			Index: 2,
			Item:  "sensor.temperature",
			Err:   errors.New("connection timeout"),
		}
		assert.Equal(t, "item 2 (sensor.temperature): connection timeout", err.Error())
	})

	t.Run("unwraps to underlying error", func(t *testing.T) {
		t.Parallel()
		underlying := errors.New("underlying error")
		err := &BatchError{
			Index: 0,
			Item:  "item",
			Err:   underlying,
		}
		assert.Equal(t, underlying, errors.Unwrap(err))
	})
}

func TestBatchErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty errors", func(t *testing.T) {
		t.Parallel()
		errs := &BatchErrors{}
		assert.False(t, errs.HasErrors())
		assert.Equal(t, "no errors", errs.Error())
	})

	t.Run("single error", func(t *testing.T) {
		t.Parallel()
		errs := &BatchErrors{}
		errs.Add(0, "entity1", errors.New("failed"))
		assert.True(t, errs.HasErrors())
		assert.Equal(t, "item 0 (entity1): failed", errs.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		t.Parallel()
		errs := &BatchErrors{}
		errs.Add(0, "entity1", errors.New("error 1"))
		errs.Add(2, "entity3", errors.New("error 2"))
		assert.True(t, errs.HasErrors())
		assert.Contains(t, errs.Error(), "2 errors occurred")
		assert.Contains(t, errs.Error(), "entity1")
		assert.Contains(t, errs.Error(), "entity3")
	})
}

func TestBatchResults(t *testing.T) {
	t.Parallel()

	t.Run("add results thread-safe", func(t *testing.T) {
		t.Parallel()
		results := &BatchResults[int]{}
		var wg sync.WaitGroup

		for i := range 100 {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				results.Add(idx, "item", idx*10, nil)
			}(i)
		}

		wg.Wait()
		assert.Len(t, results.Results, 100)
	})

	t.Run("successful filters correctly", func(t *testing.T) {
		t.Parallel()
		results := &BatchResults[string]{}
		results.Add(0, "a", "result-a", nil)
		results.Add(1, "b", "", errors.New("failed"))
		results.Add(2, "c", "result-c", nil)

		successful := results.Successful()
		assert.Len(t, successful, 2)
		assert.Equal(t, "result-a", successful[0].Result)
	})

	t.Run("failed filters correctly", func(t *testing.T) {
		t.Parallel()
		results := &BatchResults[string]{}
		results.Add(0, "a", "result-a", nil)
		results.Add(1, "b", "", errors.New("failed"))
		results.Add(2, "c", "", errors.New("also failed"))

		failed := results.Failed()
		assert.Len(t, failed, 2)
	})

	t.Run("errors returns BatchErrors", func(t *testing.T) {
		t.Parallel()
		results := &BatchResults[int]{}
		results.Add(0, "ok", 1, nil)
		results.Add(1, "fail", 0, errors.New("oops"))

		errs := results.Errors()
		assert.True(t, errs.HasErrors())
		assert.Len(t, errs.Errors, 1)
	})
}

func TestBatchConfig(t *testing.T) {
	t.Parallel()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()
		cfg := DefaultBatchConfig()
		assert.Equal(t, 0, cfg.MaxConcurrency)
		assert.True(t, cfg.ContinueOnError)
	})
}

func TestBatchExecutor(t *testing.T) {
	t.Parallel()

	t.Run("empty items returns empty results", func(t *testing.T) {
		t.Parallel()
		results := BatchExecutor(
			context.Background(),
			[]string{},
			func(s string) string { return s },
			func(_ context.Context, _ int, _ string) (int, error) {
				return 0, nil
			},
			DefaultBatchConfig(),
		)
		assert.Empty(t, results.Results)
	})

	t.Run("processes all items successfully", func(t *testing.T) {
		t.Parallel()
		items := []string{"a", "b", "c", "d", "e"}

		results := BatchExecutor(
			context.Background(),
			items,
			func(s string) string { return s },
			func(_ context.Context, _ int, item string) (string, error) {
				return item + "-processed", nil
			},
			DefaultBatchConfig(),
		)

		assert.Len(t, results.Results, 5)
		assert.Len(t, results.Successful(), 5)
		assert.Empty(t, results.Failed())
	})

	t.Run("executes concurrently", func(t *testing.T) {
		t.Parallel()
		items := []int{1, 2, 3, 4, 5}
		var concurrent atomic.Int32
		var maxConcurrent atomic.Int32

		results := BatchExecutor(
			context.Background(),
			items,
			func(_ int) string { return "" },
			func(_ context.Context, _ int, _ int) (int, error) {
				curr := concurrent.Add(1)
				// Track max concurrent executions
				for {
					currMax := maxConcurrent.Load()
					if curr <= currMax || maxConcurrent.CompareAndSwap(currMax, curr) {
						break
					}
				}
				time.Sleep(50 * time.Millisecond)
				concurrent.Add(-1)
				return 0, nil
			},
			DefaultBatchConfig(),
		)

		assert.Len(t, results.Results, 5)
		// With default config (no limit), all should run concurrently
		assert.GreaterOrEqual(t, maxConcurrent.Load(), int32(3))
	})

	t.Run("respects max concurrency", func(t *testing.T) {
		t.Parallel()
		items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		var concurrent atomic.Int32
		var maxConcurrent atomic.Int32

		results := BatchExecutor(
			context.Background(),
			items,
			func(_ int) string { return "" },
			func(_ context.Context, _ int, _ int) (int, error) {
				curr := concurrent.Add(1)
				for {
					currMax := maxConcurrent.Load()
					if curr <= currMax || maxConcurrent.CompareAndSwap(currMax, curr) {
						break
					}
				}
				time.Sleep(30 * time.Millisecond)
				concurrent.Add(-1)
				return 0, nil
			},
			BatchConfig{
				MaxConcurrency:  3,
				ContinueOnError: true,
			},
		)

		assert.Len(t, results.Results, 10)
		// Should never exceed 3 concurrent
		assert.LessOrEqual(t, maxConcurrent.Load(), int32(3))
	})

	t.Run("continues on error when configured", func(t *testing.T) {
		t.Parallel()
		items := []string{"ok1", "fail", "ok2", "fail2", "ok3"}

		results := BatchExecutor(
			context.Background(),
			items,
			func(s string) string { return s },
			func(_ context.Context, _ int, item string) (string, error) {
				if item == "fail" || item == "fail2" {
					return "", errors.New("intentional error")
				}
				return item + "-done", nil
			},
			BatchConfig{
				MaxConcurrency:  1, // Sequential for predictable order
				ContinueOnError: true,
			},
		)

		assert.Len(t, results.Results, 5)
		assert.Len(t, results.Successful(), 3)
		assert.Len(t, results.Failed(), 2)
	})

	t.Run("stops on first error when not continuing", func(t *testing.T) {
		t.Parallel()
		items := []string{"ok1", "ok2", "fail", "ok3", "ok4"}
		var processedCount atomic.Int32

		results := BatchExecutor(
			context.Background(),
			items,
			func(s string) string { return s },
			func(_ context.Context, _ int, item string) (string, error) {
				processedCount.Add(1)
				if item == "fail" {
					return "", errors.New("stop here")
				}
				time.Sleep(10 * time.Millisecond) // Ensure order
				return item, nil
			},
			BatchConfig{
				MaxConcurrency:  1, // Sequential to control order
				ContinueOnError: false,
			},
		)

		// Should have stopped after the error
		assert.Less(t, len(results.Results), 5)
		assert.GreaterOrEqual(t, len(results.Failed()), 1)
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		t.Parallel()
		items := []int{1, 2, 3, 4, 5}
		ctx, cancel := context.WithCancel(context.Background())
		var startedCount atomic.Int32

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		results := BatchExecutor(
			ctx,
			items,
			func(_ int) string { return "" },
			func(innerCtx context.Context, _ int, _ int) (int, error) {
				startedCount.Add(1)
				select {
				case <-time.After(200 * time.Millisecond):
					return 1, nil
				case <-innerCtx.Done():
					return 0, innerCtx.Err()
				}
			},
			BatchConfig{
				MaxConcurrency:  2,
				ContinueOnError: true,
			},
		)

		// Some should have been processed or started
		assert.NotEmpty(t, results.Results)
	})

	t.Run("item stringer is used for error reporting", func(t *testing.T) {
		t.Parallel()
		type CustomItem struct {
			ID   int
			Name string
		}
		items := []CustomItem{{ID: 1, Name: "first"}, {ID: 2, Name: "second"}}

		results := BatchExecutor(
			context.Background(),
			items,
			func(c CustomItem) string { return c.Name },
			func(_ context.Context, _ int, item CustomItem) (string, error) {
				if item.ID == 2 {
					return "", errors.New("failed")
				}
				return "ok", nil
			},
			DefaultBatchConfig(),
		)

		failed := results.Failed()
		require.Len(t, failed, 1)
		assert.Equal(t, "second", failed[0].Item)
	})
}

func TestExecuteBatch(t *testing.T) {
	t.Parallel()

	t.Run("uses default config", func(t *testing.T) {
		t.Parallel()
		items := []string{"a", "b", "c"}

		results := ExecuteBatch(
			context.Background(),
			items,
			func(s string) string { return s },
			func(_ context.Context, _ int, item string) (int, error) {
				return len(item), nil
			},
		)

		assert.Len(t, results.Results, 3)
		assert.Len(t, results.Successful(), 3)
	})
}

func TestExecuteBatchVoid(t *testing.T) {
	t.Parallel()

	t.Run("tracks errors for void operations", func(t *testing.T) {
		t.Parallel()
		items := []string{"ok", "fail", "ok2"}
		var processedItems sync.Map

		errs := ExecuteBatchVoid(
			context.Background(),
			items,
			func(_ context.Context, _ int, item string) error {
				processedItems.Store(item, true)
				if item == "fail" {
					return errors.New("failed to process")
				}
				return nil
			},
			DefaultBatchConfig(),
		)

		assert.True(t, errs.HasErrors())
		assert.Len(t, errs.Errors, 1)
		assert.Equal(t, "fail", errs.Errors[0].Item)

		// All items should have been processed
		count := 0
		processedItems.Range(func(_, _ any) bool {
			count++
			return true
		})
		assert.Equal(t, 3, count)
	})

	t.Run("no errors for successful operations", func(t *testing.T) {
		t.Parallel()
		items := []string{"a", "b", "c"}

		errs := ExecuteBatchVoid(
			context.Background(),
			items,
			func(_ context.Context, _ int, _ string) error {
				return nil
			},
			DefaultBatchConfig(),
		)

		assert.False(t, errs.HasErrors())
	})
}

// Benchmarks

func BenchmarkBatchExecutor_Sequential(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for range b.N {
		BatchExecutor(
			context.Background(),
			items,
			func(_ int) string { return "" },
			func(_ context.Context, _ int, item int) (int, error) {
				return item * 2, nil
			},
			BatchConfig{
				MaxConcurrency:  1,
				ContinueOnError: true,
			},
		)
	}
}

func BenchmarkBatchExecutor_Concurrent(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for range b.N {
		BatchExecutor(
			context.Background(),
			items,
			func(_ int) string { return "" },
			func(_ context.Context, _ int, item int) (int, error) {
				return item * 2, nil
			},
			DefaultBatchConfig(),
		)
	}
}

func BenchmarkBatchExecutor_LimitedConcurrency(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	b.ResetTimer()
	for range b.N {
		BatchExecutor(
			context.Background(),
			items,
			func(_ int) string { return "" },
			func(_ context.Context, _ int, item int) (int, error) {
				return item * 2, nil
			},
			BatchConfig{
				MaxConcurrency:  10,
				ContinueOnError: true,
			},
		)
	}
}
