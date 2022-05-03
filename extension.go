package lapis

// Extension interface is the base interface used for extensions
type Extension interface {
	Name() string    // The extension name
	Version() string // The extension version
}

// Extensions that hook on repository initialization
// If error is returned, then the repository initialization will be stopped and returns an error
type InitializationHookExtension[TKey comparable, TValue any] interface {
	InitializationHook(layers []Layer[TKey, TValue]) error
}

// Extensions that hook before a batched data load
// TODO: If an error is returned for a particular index, the load operation will be blocked for that index
type PreLoadHookExtension[TKey comparable, TValue any] interface {
	PreLoadHook(traceID uint64, keys []TKey) []error
}

// Extensions that hook after a batched data load
// TODO: If an error is returned for a particular index, the value for that index will not be returned into the operation caller
type PostLoadHookExtension[TKey comparable, TValue any] interface {
	PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error
}

// Extensions that hook before a batched data load from a layer
// TODO: If an error is returned for a particular index, the layer won't load that particular index
type LayerPreLoadHookExtension[TKey comparable, TValue any] interface {
	LayerPreLoadHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey) []error
}

// Extensions that hook after a batched data load from a layer
// TODO: If an error is returned for a particular index, the value resolved by this layer won't be combined with the result, the next layer will try to resolve it
type LayerPostLoadHookExtension[TKey comparable, TValue any] interface {
	LayerPostLoadHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) []error
}

// Extensions that hook before a data set operation
// TODO: If an error is returned for a particular index, the set operation will be blocked for that index
type PreSetHookExtension[TKey comparable, TValue any] interface {
	PreSetHook(traceID uint64, keys []TKey, values []TValue) []error
}

// Extensions that hook after a data set operation
type PostSetHookExtension[TKey comparable, TValue any] interface {
	PostSetHook(traceID uint64, keys []TKey, values []TValue, errors [][]error)
}

// Extensions that hook before a data set operation of a layer
// TODO: If an error is returned for a particular index, the set operation will be blocked for that index at the current layer
type LayerPreSetHookExtension[TKey comparable, TValue any] interface {
	LayerPreSetHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey, values []TValue) []error
}

// Extensions that hook after a data set operation of a layer
type LayerPostSetHookExtension[TKey comparable, TValue any] interface {
	LayerPostSetHook(traceID uint64, layer Layer[TKey, TValue], keys []TKey, values []TValue, errors []error)
}
