package layer

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/collection/queue"
	"github.com/flowscan/lapis/collection/tuple"
)

// Configuration for the memory data layer
type MemoryConfig struct {
	// The duration of the cached data
	Retention time.Duration
}

// Memory layer is map-based in-memory cache, it should be used as the first line of cache
// with short data expiration
type Memory[TKey comparable, TValue any] struct {
	config            MemoryConfig
	data              map[TKey]TValue
	mu                sync.RWMutex
	invalidationQueue *queue.Queue[tuple.Pair[time.Time, TKey]]
}

// Unique identifier for this layer used for logging and metric purposes
func (l *Memory[TKey, TValue]) Identifier() string { return "memory" }

// The function that will be used to resolve a set of keys
func (l *Memory[TKey, TValue]) Get(keys []TKey) ([]TValue, []error) {
	result := make([]TValue, len(keys))
	errors := make([]error, len(keys))
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i, k := range keys {
		if v, ok := l.data[k]; ok {
			result[i] = v
		} else {
			errors[i] = lapis.NewErrNotFound(k)
		}
	}
	return result, errors
}

// The function that will be called for successful resolvers
func (l *Memory[TKey, TValue]) Set(keys []TKey, values []TValue) []error {
	l.mu.Lock()
	for i, k := range keys {
		l.data[k] = values[i]
		l.invalidationQueue.Enqueue(tuple.NewPair(time.Now().Add(l.config.Retention), k))
	}
	l.mu.Unlock()
	return nil
}

// Create a new in-memory data layer
func NewMemory[TKey comparable, TValue any](config MemoryConfig) *Memory[TKey, TValue] {
	l := &Memory[TKey, TValue]{
		config: config,
		data:   make(map[TKey]TValue),
	}
	if config.Retention > 0 {
		l.startInvalidator()
	}
	return l
}

// deletes records on the cache
func (l *Memory[TKey, TValue]) startInvalidator() {
	l.invalidationQueue = queue.NewQueue[tuple.Pair[time.Time, TKey]](1)
	go func() {
		throttle := newThrottler(l.config.Retention)
		for {
			throttle.Throttle()
			for l.invalidationQueue.Len() > 0 {
				nextJob := l.invalidationQueue.Dequeue()
				if time.Now().Before(nextJob.V1) {
					time.Sleep(time.Until(nextJob.V1))
				}
				l.mu.Lock()
				delete(l.data, nextJob.V2)
				l.mu.Unlock()
			}
		}
	}()
}

type throttler struct {
	lastInvoke   time.Time
	throttleTime time.Duration
}

func (t *throttler) Throttle() {
	timePassed := time.Since(t.lastInvoke)
	if timePassed < t.throttleTime {
		time.Sleep(t.throttleTime - timePassed)
	}
	t.lastInvoke = time.Now()
}

func newThrottler(throttleTime time.Duration) *throttler {
	return &throttler{
		lastInvoke:   time.Now(),
		throttleTime: throttleTime,
	}
}
