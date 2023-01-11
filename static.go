package lapis

// a static repository doesn't have any keys, therefore we need to mock the key in the static repository API
// placeholder type for the key used in static repositories
type staticTypePlaceholder struct{}

// create an array of static key of size 1
func staticKey() []staticTypePlaceholder {
	return []staticTypePlaceholder{staticTypePlaceholder{}}
}

// StaticRepository is a special type of repository where a key or input
// is not required to fetch data from the repository
type StaticRepository[TValue any] struct {
	repo Repository[staticTypePlaceholder, TValue]
}

// Get data from the repository
func (r *StaticRepository[TValue]) Get(value TValue, options ...LoadFlag) (TValue, error) {
	return r.repo.Load(staticTypePlaceholder{}, options...)
}

// Set the repository data to all of the layers
// Returns an array of errors with each item represents an error returned by a layer
func (r *StaticRepository[TValue]) Set(value TValue, options ...SetOption) []error {
	return kvToStaticErrors(r.repo.SetAll(staticKey(), []TValue{value}, options...))
}

// take the 2D array result from KV repository implementation, then restructure it for the static repository implementation
func kvToStaticErrors(original [][]error) []error {
	result := make([]error, len(original))
	for i, layerValues := range original {
		if len(layerValues) > 0 && layerValues[0] != nil {
			result[i] = layerValues[0]
		}
	}
	return result
}
