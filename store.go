package lapis

import (
	"sync/atomic"
	"time"
)

type Store[TKey comparable, TValue any] struct {
	// identifier for the store
	identifier string

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

// Get the identifier of the store
func (r *Store[TKey, TValue]) Identifier() string {
	return r.identifier
}

func (r *Store[TKey, TValue]) getTraceID() uint64 {
	return atomic.AddUint64(&r.traceCounter, 1)
}

// Create a new data store with the given configuration
func New[TKey comparable, TValue any](config Config[TKey, TValue]) (*Store[TKey, TValue], error) {
	r := &Store[TKey, TValue]{
		layers:     config.Layers,
		identifier: config.Identifier,
	}
	if config.Batcher.MaxBatch > 0 {
		r.useBatcher = true
		r.batcher = Batcher[TKey, TValue]{
			resolver: r.resolve,
			wait:     zeroFallback(config.Batcher.Wait, 1*time.Millisecond),
			maxBatch: zeroFallback(config.Batcher.MaxBatch, 256),
			batches:  make(map[TKey]*batch[TKey, TValue]),
		}
	}

	r.registerExtensions(config.Extensions)

	// Execute initialization hooks
	for _, hook := range r.initializationHooks {
		err := hook.InitializationHook(r, config.Layers)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (r *Store[TKey, TValue]) registerExtensions(extensions []Extension) {
	r.initializationHooks = make([]InitializationHookExtension[TKey, TValue], 0)
	r.preLoadHooks = make([]PreLoadHookExtension[TKey, TValue], 0)
	r.postLoadHooks = make([]PostLoadHookExtension[TKey, TValue], 0)
	r.layerPreLoadHooks = make([]LayerPreLoadHookExtension[TKey, TValue], 0)
	r.layerPostLoadHooks = make([]LayerPostLoadHookExtension[TKey, TValue], 0)
	r.preSetHooks = make([]PreSetHookExtension[TKey, TValue], 0)
	r.postSetHooks = make([]PostSetHookExtension[TKey, TValue], 0)
	r.layerPreSetHooks = make([]LayerPreSetHookExtension[TKey, TValue], 0)
	r.layerPostSetHooks = make([]LayerPostSetHookExtension[TKey, TValue], 0)
	for _, ext := range extensions {
		if ext, ok := ext.(InitializationHookExtension[TKey, TValue]); ok {
			r.initializationHooks = append(r.initializationHooks, ext)
		}
		if ext, ok := ext.(PreLoadHookExtension[TKey, TValue]); ok {
			r.preLoadHooks = append(r.preLoadHooks, ext)
		}
		if ext, ok := ext.(PostLoadHookExtension[TKey, TValue]); ok {
			r.postLoadHooks = append(r.postLoadHooks, ext)
		}
		if ext, ok := ext.(LayerPreLoadHookExtension[TKey, TValue]); ok {
			r.layerPreLoadHooks = append(r.layerPreLoadHooks, ext)
		}
		if ext, ok := ext.(LayerPostLoadHookExtension[TKey, TValue]); ok {
			r.layerPostLoadHooks = append(r.layerPostLoadHooks, ext)
		}
		if ext, ok := ext.(PreSetHookExtension[TKey, TValue]); ok {
			r.preSetHooks = append(r.preSetHooks, ext)
		}
		if ext, ok := ext.(PostSetHookExtension[TKey, TValue]); ok {
			r.postSetHooks = append(r.postSetHooks, ext)
		}
		if ext, ok := ext.(LayerPreSetHookExtension[TKey, TValue]); ok {
			r.layerPreSetHooks = append(r.layerPreSetHooks, ext)
		}
		if ext, ok := ext.(LayerPostSetHookExtension[TKey, TValue]); ok {
			r.layerPostSetHooks = append(r.layerPostSetHooks, ext)
		}
	}
}
