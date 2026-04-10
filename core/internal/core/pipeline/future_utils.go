package pipeline

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Future represents an async computation result
type Future[T any] struct {
	result T
	err    error
	done   chan struct{}
}

// NewFuture creates a new Future
func NewFuture[T any]() *Future[T] {
	return &Future[T]{
		done: make(chan struct{}),
	}
	}

// SetResult sets the result and marks the future as done
func (f *Future[T]) SetResult(result T, err error) {
	f.result = result
	f.err = err
	close(f.done)
}

// Get waits for and returns the result
func (f *Future[T]) Get(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.result, f.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// IsDone returns true if the future is complete
func (f *Future[T]) IsDone() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Zip waits for all futures to complete and returns their results
// If any future fails, returns the first error encountered
func Zip[T any](futures []*Future[T]) ([]T, error) {
	if len(futures) == 0 {
		return []T{}, nil
	}

	results := make([]T, len(futures))
	var mu sync.Mutex
	var firstErr error

	var wg sync.WaitGroup
	wg.Add(len(futures))

	for i, f := range futures {
		go func(index int, future *Future[T]) {
			defer wg.Done()
			
			result, err := future.Get(context.Background())
			
			mu.Lock()
			defer mu.Unlock()
			
			results[index] = result
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}(i, f)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

// ZipWithContext waits for all futures to complete with context cancellation support
func ZipWithContext[T any](ctx context.Context, futures []*Future[T]) ([]T, error) {
	if len(futures) == 0 {
		return []T{}, nil
	}

	g, ctx := errgroup.WithContext(ctx)
	results := make([]T, len(futures))

	for i, f := range futures {
		i, f := i, f // capture loop variables
		g.Go(func() error {
			result, err := f.Get(ctx)
			if err != nil {
				return err
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteAsync executes a function asynchronously and returns a Future
func ExecuteAsync[T any](fn func() (T, error)) *Future[T] {
	future := NewFuture[T]()
	
	go func() {
		result, err := fn()
		future.SetResult(result, err)
	}()
	
	return future
}

// ExecuteAsyncWithContext executes a function asynchronously with context support
func ExecuteAsyncWithContext[T any](ctx context.Context, fn func(context.Context) (T, error)) *Future[T] {
	future := NewFuture[T]()
	
	go func() {
		result, err := fn(ctx)
		future.SetResult(result, err)
	}()
	
	return future
}

// WaitForAll waits for all futures to complete, ignoring errors
func WaitForAll[T any](futures []*Future[T]) {
	var wg sync.WaitGroup
	wg.Add(len(futures))
	
	for _, f := range futures {
		go func(future *Future[T]) {
			defer wg.Done()
			_, _ = future.Get(context.Background())
		}(f)
	}
	
	wg.Wait()
}

// AnyDone returns true if any future is done
func AnyDone[T any](futures []*Future[T]) bool {
	for _, f := range futures {
		if f.IsDone() {
			return true
		}
	}
	return false
}

// AllDone returns true if all futures are done
func AllDone[T any](futures []*Future[T]) bool {
	for _, f := range futures {
		if !f.IsDone() {
			return false
		}
	}
	return true
}
