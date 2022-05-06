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
	resolver func(keys []TKey) ([]TValue, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *loaderBatch[TKey, TValue]

	// mutex to prevent races
	mu sync.Mutex
}

type loaderBatch[TKey comparable, TValue any] struct {
	keys    []TKey
	data    []TValue
	error   []error
	closing bool
	done    chan struct{}
}

// Load a value by key, batching and caching will be applied automatically
func (l *Batcher[TKey, TValue]) Load(key TKey) (TValue, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *Batcher[TKey, TValue]) LoadThunk(key TKey) func() (TValue, error) {
	l.mu.Lock()
	if l.batch == nil {
		l.batch = &loaderBatch[TKey, TValue]{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (TValue, error) {
		<-batch.done

		var data TValue
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
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

// LoadAllThunk returns a function that when called will block waiting for valyes
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

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *loaderBatch[TKey, TValue]) keyIndex(l *Batcher[TKey, TValue], key TKey) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.resolveBatch(l)
		}
	}

	return pos
}

func (b *loaderBatch[TKey, TValue]) startTimer(l *Batcher[TKey, TValue]) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.resolveBatch(l)
}

func (b *loaderBatch[TKey, TValue]) resolveBatch(l *Batcher[TKey, TValue]) {
	b.data, b.error = l.resolver(b.keys)
	close(b.done)
}
