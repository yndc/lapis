package tlru

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/collection/tuple"
)

// an implementation of Time-Aware Least Recent Used in-memory cache
type Cache[TKey comparable, TValue any] struct {
	data              map[TKey]TValue
	invalidationQueue *PriorityQueue[TKey]
	config            Config
	counter           uint64
	mu                sync.RWMutex
}

// get values from the given keys
func (c *Cache[TKey, TValue]) Get(keys []TKey) ([]TValue, []error) {
	result := make([]TValue, len(keys))
	errors := make([]error, len(keys))
	c.mu.RLock()
	defer c.mu.RUnlock()
	for i, k := range keys {
		if v, ok := c.data[k]; ok {
			result[i] = v
		} else {
			errors[i] = lapis.NewErrNotFound(k)
		}
	}
	return result, errors
}

// set values for the given keys
func (l *Cache[TKey, TValue]) Set(keys []TKey, values []TValue) []error {
	l.mu.Lock()
	for i, k := range keys {
		l.data[k] = values[i]
		l.invalidationQueue.Enqueue(tuple.NewPair(time.Now().Add(l.config.Retention), k))
	}
	l.mu.Unlock()
	return nil
}

// create a new TLRU cache
func NewCache[TKey comparable, TValue any](config Config) (*Cache[TKey, TValue], error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	c := &Cache[TKey, TValue]{
		data:              make(map[TKey]TValue),
		invalidationQueue: NewQueue[TKey](),
		mu:                sync.RWMutex{},
		config:            config,
	}

	return c, nil
}
