package lapis

import (
	"sync"
	"time"
)

// Configuration for the batcher
type BatcherConfig[TKey comparable, TValue any] struct {
	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// Batcher batches and caches requests
type Batcher[TKey comparable, TValue any] struct {
	// the resolver for the batched requests
	resolver func(keys []TKey, finishKey func(index int, value TValue, err error))

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// the current unfinished batch
	pendingBatch *batch[TKey, TValue]

	// a map of keys to active batches that has the key in the batch
	batches map[TKey]*batch[TKey, TValue]

	// mutex to prevent races
	mu sync.Mutex

	// default load flags
	defaultLoadFlags LoadFlag
}

// Load a value by key, batching and caching will be applied automatically
func (l *Batcher[TKey, TValue]) Load(key TKey) (TValue, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block the thread until the requested data is resolved
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *Batcher[TKey, TValue]) LoadThunk(key TKey, flags ...LoadFlag) func() (TValue, error) {
	var currentBatch *batch[TKey, TValue]
	l.mu.Lock()

	// priority of batch to be used:
	// (1) existing batch with the same key
	// (2) pending batch
	// (3) create a new batch
	if existingBatch, ok := l.batches[key]; !hasLoadFlag(l.defaultLoadFlags, flags, LoadNoShareBatch) && ok {
		currentBatch = existingBatch
	} else {
		if l.pendingBatch != nil {
			currentBatch = l.pendingBatch
		} else {
			currentBatch = &batch[TKey, TValue]{allDone: make(chan struct{})}
			l.pendingBatch = currentBatch
		}
		l.batches[key] = currentBatch
	}

	index := currentBatch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (TValue, error) {
		<-currentBatch.done[index]

		var data TValue
		if index < len(currentBatch.data) {
			data = currentBatch.data[index]
		}

		var err error
		if len(currentBatch.errors) == 1 {
			err = currentBatch.errors[0]
		} else if currentBatch.errors != nil {
			err = currentBatch.errors[index]
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *Batcher[TKey, TValue]) LoadAll(keys []TKey) ([]TValue, []error) {
	results := make([]func() (TValue, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	result := make([]TValue, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		result[i], errors[i] = thunk()
	}
	return result, errors
}

// LoadAllThunk returns a function that when called will block waiting for values
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *Batcher[TKey, TValue]) LoadAllThunk(keys []TKey) func() ([]TValue, []error) {
	results := make([]func() (TValue, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]TValue, []error) {
		values := make([]TValue, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			values[i], errors[i] = thunk()
		}
		return values, errors
	}
}
