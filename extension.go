package lapis

// Extension interface is the base interface used for extensions
type Extension interface {
	Name() string    // The extension name
	Version() string // The extension version
}

// Extensions that hook on repository initialization
type InitializationHookExtension[TKey comparable, TValue any] interface {
	InitializationHook(layers []Layer[TKey, TValue]) error
}

// Extensions that hook before a batched data load
type PreLoadHookExtension[TKey comparable, TValue any] interface {
	PreLoadHook(traceID uint64, keys []TKey) []error
}

// Extensions that hook after a batched data load
type PostLoadHookExtension[TKey comparable, TValue any] interface {
	PreLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error
}

// Extensions that hook before a batched data load from a layer
type LayerPreLoadHookExtension[TKey comparable, TValue any] interface {
	LayerPreLoadHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey) []error
}

// Extensions that hook after a batched data load from a layer
type LayerPostLoadHookExtension[TKey comparable, TValue any] interface {
	LayerPostLoadHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) []error
}

// Extensions that hook before a data set operation
type PreSetHookExtension[TKey comparable, TValue any] interface {
	PreSetHook(traceID uint64, keys []TKey) []error
}

// Extensions that hook after a data set operation
type PostSetHookExtension[TKey comparable, TValue any] interface {
	PostSetHook(traceID uint64, keys []TKey) []error
}
