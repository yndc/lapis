package lapis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flowscan/lapis/collection/tuple"
	"github.com/flowscan/nft-analytics-go/pkg/collection/bucket"
	"github.com/flowscan/nft-analytics-go/pkg/collection/set"
)

// Option on receiver channel overflow
type OverflowBehaviour int

const (
	// Ignore if the receiver channel is full, the subscriber might miss out on some data
	OverflowBehaviourIgnore OverflowBehaviour = iota

	// Wait for the subscriber to be able to receive the data
	OverflowBehaviourWait

	// Forcefully removes and closes the subscriber channel
	OverflowBehaviourForceRemove
)

// LoaderConfig captures the config to create a new stream loader
type LoaderConfig[TKey comparable, TValue any] struct {
	// Open is a method to opens a new data stream for the loader
	Open func(key TKey, stop <-chan struct{}) (<-chan TValue, error)

	// KeepAlive is the time duration to keep a data stream alive for a particular key alive when the last subscriber disconnects
	KeepAlive time.Duration

	// AlwaysKeepAlive is a list of keys that will be opened on initialization and will be kept alive even if there's no subscribers
	AlwaysKeepAlive bool

	// AlwaysKeepAlive is a list of keys that will be opened on initialization and will be kept alive even if there's no subscribers
	AlwaysKeepAliveKeys []TKey

	// Buffer is the buffer size used for the subscribers
	Buffer int

	// OverflowBehaviour is the expected behaviour when one of the subscriber channel is full
	OverflowBehaviour OverflowBehaviour
}

// New creates a new stream loader given a fetch, wait, and maxBatch
func NewLoader[TKey comparable, TValue any](config LoaderConfig[TKey, TValue]) *Loader[TKey, TValue] {
	return &Loader[TKey, TValue]{
		opener:              config.Open,
		keepAlive:           config.KeepAlive,
		alwaysKeepAliveKeys: set.New(config.AlwaysKeepAliveKeys...),
		alwaysKeepAlive:     config.AlwaysKeepAlive,
		nonce:               bucket.NewSum[TKey, int](),
		stopper:             make(map[TKey]chan<- struct{}),
		subscribers:         bucket.NewSet[TKey, chan TValue](),
		subscriberMapper:    make(map[<-chan TValue]tuple.Pair[TKey, chan TValue]),
	}
}

// Loader batches and caches requests
type Loader[TKey comparable, TValue any] struct {
	// identifier for the store
	identifier string

	// a method to opens a new data stream for the loader
	opener func(key TKey, stop <-chan struct{}) (<-chan TValue, error)

	// KeepAlive is the time duration to keep a data stream alive for a particular key alive when the last subscriber disconnects
	keepAlive time.Duration

	// AlwaysKeepAlive if this is true, then streams won't be closed if there's no subscribers
	alwaysKeepAlive bool

	// AlwaysKeepAlive is a list of keys that will be opened on initialization and will be kept alive even if there's no subscribers
	alwaysKeepAliveKeys set.Set[TKey]

	// Buffer is the buffer size used for the subscribers
	buffer int

	// The last stream value
	lastValue *TValue

	// expected behaviour when one of the subscriber channel is full
	overflowBehaviour OverflowBehaviour

	// currently open stopper
	stopper   map[TKey]chan<- struct{}
	stopperMu sync.RWMutex

	// subscribers for each streams
	subscribers      bucket.Set[TKey, chan TValue]
	subscriberMapper map[<-chan TValue]tuple.Pair[TKey, chan TValue]
	subscribersMu    sync.RWMutex

	// nonce to help with ordering
	nonce   *bucket.Sum[TKey, int]
	nonceMu sync.RWMutex

	// data resolver layers in this store
	layers []Layer[TKey, TValue]

	// batcher if batching is enabled
	batcher Batcher[TKey, TValue]

	// trace counter for trace ID assignment
	traceCounter uint64

	// flag to use batcher
	useBatcher bool

	// default load flags
	defaultLoadFlags LoadFlag

	// hooks
	initializationHooks []InitializationHookExtension[TKey, TValue]
	preLoadHooks        []PreLoadHookExtension[TKey, TValue]
	postLoadHooks       []PostLoadHookExtension[TKey, TValue]
	layerPreLoadHooks   []LayerPreLoadHookExtension[TKey, TValue]
	layerPostLoadHooks  []LayerPostLoadHookExtension[TKey, TValue]
	preSetHooks         []PreSetHookExtension[TKey, TValue]
	postSetHooks        []PostSetHookExtension[TKey, TValue]
	layerPreSetHooks    []LayerPreSetHookExtension[TKey, TValue]
	layerPostSetHooks   []LayerPostSetHookExtension[TKey, TValue]
}

// Subscribe to a stream by key, if existing stream already exists it will be used without creating a new one
func (l *Loader[TKey, TValue]) Subscribe(key TKey) (<-chan TValue, error) {
	c := make(chan TValue, l.buffer)

	l.stopperMu.RLock()

	// open a new stream if not exists
	if _, ok := l.stopper[key]; !ok {
		l.stopperMu.RUnlock()
		stop := make(chan struct{})
		stream, err := l.opener(key, stop)
		if err != nil {
			return nil, err
		}

		l.stopperMu.Lock()
		l.nonceMu.Lock()

		l.stopper[key] = stop
		l.nonce.Increment(key)

		l.stopperMu.Unlock()
		l.nonceMu.Unlock()
		go l.fanOut(key, stream)
	}

	l.registerSubscriber(c, key)
	return c, nil
}

// Unsubscribe to a stream by key, if there are no subscribers left for this stream,
// the stream will be closed unless it is configured to be kept alive
func (l *Loader[TKey, TValue]) Unsubscribe(c <-chan TValue) error {
	l.subscribersMu.Lock()
	defer l.subscribersMu.Unlock()

	if original, ok := l.subscriberMapper[c]; ok {
		remaining := l.subscribers.Remove(original.V1, original.V2)
		if remaining == 0 {
			l.closeStream(original.V1)
		}
		delete(l.subscriberMapper, c)
	} else {
		return ErrNotFound(fmt.Errorf("subscriber not found"))
	}
	return nil
}

// Subscribe to a stream by key, will automatically close the channel after the context is done
func (l *Loader[TKey, TValue]) SubscribeCtx(ctx context.Context, key TKey) (<-chan TValue, error) {
	c, err := l.Subscribe(key)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		l.Unsubscribe(c)
	}()

	return c, nil
}

func (l *Loader[TKey, TValue]) registerSubscriber(c chan TValue, key TKey) {
	l.subscribersMu.Lock()
	l.subscribers.Add(key, c)
	l.subscriberMapper[c] = tuple.NewPair(key, c)
	l.subscribersMu.Unlock()
}

func (l *Loader[TKey, TValue]) closeStream(key TKey) {
	if l.alwaysKeepAlive || l.alwaysKeepAliveKeys.Has(key) {
		return
	}
	if l.keepAlive > 0 {
		l.nonceMu.RLock()
		nonce := l.nonce.Get(key)
		l.nonceMu.RUnlock()
		go func() {
			time.Sleep(l.keepAlive)
			l.nonceMu.RLock()
			if l.nonce.Get(key) != nonce {
				defer l.nonceMu.RUnlock()
				return
			}
			l.nonceMu.RUnlock()
			l.stopperMu.RLock()
			l.stopper[key] <- struct{}{}
			l.stopperMu.RUnlock()
		}()
	} else {
		l.stopperMu.RLock()
		l.stopper[key] <- struct{}{}
		l.stopperMu.RUnlock()
	}
}

func (l *Loader[TKey, TValue]) fanOut(key TKey, source <-chan TValue) {
	for {
		v := <-source
		dsts := l.subscribers.Get(key)
		for dst := range dsts {
			select {
			case dst <- v:
			default:
				switch l.overflowBehaviour {
				case OverflowBehaviourWait:
					dst <- v
				case OverflowBehaviourForceRemove:
					dsts.Remove(dst)
					close(dst)
				default:
					continue
				}
			}
		}
	}
}
