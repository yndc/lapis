package lapis

// Configuration for a repository
type Config[TKey comparable, TValue any] struct {
	// Identifier for this repository
	Identifier string

	// Configuration for the batcher, if not included batching will be disabled
	Batcher *BatcherConfig[TKey, TValue]

	// The data resolver layers for this repository, executed from the first to the last
	Layers []Layer[TKey, TValue]

	// Default load flags
	DefaultLoadFlags LoadFlag

	// Default set flags
	DefaultSetFlags int

	// Array of extensions to be used
	Extensions []Extension
}
