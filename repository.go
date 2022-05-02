package lapis

import "sync/atomic"

type Repository[TKey comparable, TValue any] struct {
	// data resolver layers in this repository
	layers []Layer[TKey, TValue]

	// load a data in thunk
	loadThunk func(key TKey) func() (TValue, error)

	// batcher if batching is enabled
	batcher Batcher[TKey, TValue]

	// trace counter for trace ID assignment
	traceCounter uint64

	// hooks
	initializationHooks []InitializationHookExtension[TKey, TValue]
}

func (r *Repository[TKey, TValue]) getTraceID() uint64 {
	return atomic.AddUint64(&r.traceCounter, 1)
}

// Create a new data repository with the given configuration
func New[TKey comparable, TValue any](config Config[TKey, TValue]) (*Repository[TKey, TValue], error) {
	repository := &Repository[TKey, TValue]{}
	if config.Batcher.MaxBatch > 0 {
		repository.batcher = Batcher[TKey, TValue]{}
	}

	// execute the hooks
	// for _, hook := range

	return repository, nil
}
