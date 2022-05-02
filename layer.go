package lapis

// Handler is any function that takes a model key and will return the model value
type Handler[TKey comparable, TValue any] func(key TKey) (TValue, error)

// BatchHandler is any function that takes a list of model keys and will return the values
type BatchHandler[TKey comparable, TValue any] func(keys []TKey) ([]TValue, []error)

// Layer is an interface for data layers, they take the data key and will return the result if available
// If the resolver function returns null, for a particular data key, the key will be given into the next
// data layer to be resolved
type Layer[TKey comparable, TValue any] interface {
	// Unique identifier for this layer used for logging and metric purposes
	Identifier() string

	// The function that will be called to load values from the given set of keys
	Get(keys []TKey) ([]TValue, []error)

	// The function that will be called for setting data to be primed that is resolved by the layers after this
	Set(keys []TKey, values []TValue) []error
}
