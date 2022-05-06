package lapis

import (
	"sync"
	"time"
)

// Convert a handler into a batch handler, each keys will be handled in parallel
func Batchify[TKey comparable, TValue any](f Handler[TKey, TValue]) BatchHandler[TKey, TValue] {
	return func(keys []TKey) ([]TValue, []error) {
		values := make([]TValue, len(keys))
		errors := make([]error, len(keys))
		wg := sync.WaitGroup{}
		for i := range keys {
			wg.Add(1)
			capturedIndex := i
			go func() {
				defer wg.Done()
				value, err := f(keys[capturedIndex])
				if err != nil {
					errors[capturedIndex] = err
				} else {
					values[capturedIndex] = value
				}
			}()
		}
		wg.Wait()
		return values, errors
	}
}

// Convert a batch handler to a single handler
func Singlify[TKey comparable, TValue any](f BatchHandler[TKey, TValue]) Handler[TKey, TValue] {
	return func(key TKey) (TValue, error) {
		keys := []TKey{key}
		result, errors := f(keys)
		if len(errors) > 0 && errors[0] != nil {
			return zero[TValue](), errors[0]
		}
		return result[0], nil
	}
}

// Return the zero value of the given generic type
func zero[T any]() T {
	var zero T
	return zero
}

func hasFlag(flags []int, flag int) bool {
	sum := 0
	for _, f := range flags {
		sum = sum | f
	}
	return (sum & flag) == flag
}

func zeroFallback[T comparable](input T, fallback T) T {
	var zero T
	if input == zero {
		return fallback
	}
	return input
}

// stopwatch measures the time elapsed between the start of the function
func stopwatch() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		elapsed := time.Since(start)
		start = time.Now()
		return elapsed
	}
}
