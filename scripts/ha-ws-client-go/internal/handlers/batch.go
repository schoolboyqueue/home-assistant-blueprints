// Package handlers provides command handlers for the CLI.
package handlers

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// BatchError represents an error that occurred during batch processing.
// It tracks which item failed and the associated error.
type BatchError struct {
	Index int    // Index of the failed item in the original slice
	Item  string // String representation of the item (for error messages)
	Err   error  // The underlying error
}

func (e *BatchError) Error() string {
	return fmt.Sprintf("item %d (%s): %v", e.Index, e.Item, e.Err)
}

func (e *BatchError) Unwrap() error {
	return e.Err
}

// BatchErrors is a collection of errors from batch processing.
type BatchErrors struct {
	Errors []*BatchError
}

func (e *BatchErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors occurred:\n", len(e.Errors))
	for _, err := range e.Errors {
		fmt.Fprintf(&sb, "  - %s\n", err.Error())
	}
	return sb.String()
}

// HasErrors returns true if there are any errors in the collection.
func (e *BatchErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Add adds a new error to the collection (thread-safe).
func (e *BatchErrors) Add(index int, item string, err error) {
	e.Errors = append(e.Errors, &BatchError{
		Index: index,
		Item:  item,
		Err:   err,
	})
}

// BatchResult represents the result of processing a single item.
type BatchResult[T any] struct {
	Index  int    // Index in the original slice
	Item   string // String representation of the item
	Result T      // The result (zero value if error)
	Err    error  // Error if processing failed
}

// BatchResults holds all results from batch processing.
type BatchResults[T any] struct {
	Results []*BatchResult[T]
	mu      sync.Mutex
}

// Add adds a result to the collection (thread-safe).
func (r *BatchResults[T]) Add(index int, item string, result T, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Results = append(r.Results, &BatchResult[T]{
		Index:  index,
		Item:   item,
		Result: result,
		Err:    err,
	})
}

// Successful returns only the successful results.
func (r *BatchResults[T]) Successful() []*BatchResult[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	var successful []*BatchResult[T]
	for _, res := range r.Results {
		if res.Err == nil {
			successful = append(successful, res)
		}
	}
	return successful
}

// Failed returns only the failed results.
func (r *BatchResults[T]) Failed() []*BatchResult[T] {
	r.mu.Lock()
	defer r.mu.Unlock()
	var failed []*BatchResult[T]
	for _, res := range r.Results {
		if res.Err != nil {
			failed = append(failed, res)
		}
	}
	return failed
}

// Errors returns a BatchErrors containing all errors.
func (r *BatchResults[T]) Errors() *BatchErrors {
	r.mu.Lock()
	defer r.mu.Unlock()
	errs := &BatchErrors{}
	for _, res := range r.Results {
		if res.Err != nil {
			errs.Add(res.Index, res.Item, res.Err)
		}
	}
	return errs
}

// BatchConfig configures batch execution behavior.
type BatchConfig struct {
	// MaxConcurrency limits the number of concurrent operations.
	// If 0, defaults to len(items) (no limit).
	MaxConcurrency int

	// ContinueOnError determines if processing should continue after an error.
	// If true, all items are processed and errors are collected.
	// If false, processing stops on the first error.
	ContinueOnError bool
}

// DefaultBatchConfig returns the default batch configuration.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		MaxConcurrency:  0, // No limit (all items processed concurrently)
		ContinueOnError: true,
	}
}

// BatchExecutor executes a function for each item in a slice concurrently.
// It returns results for all items, including errors.
//
// Type parameters:
//   - I: Input item type
//   - O: Output result type
//
// The processor function receives the context, item index, and item value.
// It should return the result and any error.
func BatchExecutor[I any, O any](
	ctx context.Context,
	items []I,
	itemStringer func(I) string,
	processor func(ctx context.Context, index int, item I) (O, error),
	config BatchConfig,
) *BatchResults[O] {
	results := &BatchResults[O]{
		Results: make([]*BatchResult[O], 0, len(items)),
	}

	if len(items) == 0 {
		return results
	}

	// Determine concurrency
	concurrency := config.MaxConcurrency
	if concurrency <= 0 || concurrency > len(items) {
		concurrency = len(items)
	}

	// Create a semaphore channel to limit concurrency
	sem := make(chan struct{}, concurrency)

	// Create a cancelable context if we need to stop on error
	execCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

itemLoop:
	for i, item := range items {
		// Check if context is already canceled
		if execCtx.Err() != nil {
			break itemLoop
		}

		// Check for early termination due to error
		errMu.Lock()
		shouldStop := !config.ContinueOnError && firstErr != nil
		errMu.Unlock()
		if shouldStop {
			break itemLoop
		}

		// Try to acquire semaphore with context cancellation check
		select {
		case sem <- struct{}{}:
			// Acquired semaphore, continue
		case <-execCtx.Done():
			// Context canceled while waiting for semaphore
			break itemLoop
		}

		wg.Add(1)
		go func(idx int, itm I) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			itemStr := ""
			if itemStringer != nil {
				itemStr = itemStringer(itm)
			}

			result, err := processor(execCtx, idx, itm)
			results.Add(idx, itemStr, result, err)

			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				if !config.ContinueOnError {
					cancel()
				}
				errMu.Unlock()
			}
		}(i, item)
	}

	wg.Wait()
	return results
}

// ExecuteBatch is a simplified version of BatchExecutor that uses default config.
func ExecuteBatch[I any, O any](
	ctx context.Context,
	items []I,
	itemStringer func(I) string,
	processor func(ctx context.Context, index int, item I) (O, error),
) *BatchResults[O] {
	return BatchExecutor(ctx, items, itemStringer, processor, DefaultBatchConfig())
}

// ExecuteBatchVoid is for batch operations that don't return a value.
// It's useful for operations like subscriptions or cleanup.
func ExecuteBatchVoid(
	ctx context.Context,
	items []string,
	processor func(ctx context.Context, index int, item string) error,
	config BatchConfig,
) *BatchErrors {
	results := BatchExecutor(ctx, items, func(s string) string { return s },
		func(ctx context.Context, index int, item string) (struct{}, error) {
			return struct{}{}, processor(ctx, index, item)
		}, config)

	return results.Errors()
}
