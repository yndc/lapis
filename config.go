package lapis

// Configuration for a repository
type Config[TKey comparable, TValue any] struct {

	// Configuration for the batcher, if not included batching will be disabled
	Batcher BatcherConfig[TKey, TValue]

	// The data resolver layers for this repository, executed from the first to the last
	Layers []Layer[TKey, TValue]
}
