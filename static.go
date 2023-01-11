package lapis

// a static store doesn't have any keys, therefore we need to mock the key in the static store API
// placeholder type for the key used in static stores
type staticTypePlaceholder struct{}

// create an array of static key of size 1
func staticKey() []staticTypePlaceholder {
	return []staticTypePlaceholder{{}}
}

// StaticStore is a special type of store where a key or input
// is not required to fetch data from the store
type StaticStore[TValue any] struct {
	store Store[staticTypePlaceholder, TValue]
}

// Get data from the store
func (r *StaticStore[TValue]) Get(value TValue, options ...LoadFlag) (TValue, error) {
	return r.store.Load(staticTypePlaceholder{}, options...)
}

// Set the store data to all of the layers
// Returns an array of errors with each item represents an error returned by a layer
func (r *StaticStore[TValue]) Set(value TValue, options ...SetOption) []error {
	return kvToStaticErrors(r.store.SetAll(staticKey(), []TValue{value}, options...))
}

// take the 2D array result from KV store implementation, then restructure it for the static store implementation
func kvToStaticErrors(original [][]error) []error {
	result := make([]error, len(original))
	for i, layerValues := range original {
		if len(layerValues) > 0 && layerValues[0] != nil {
			result[i] = layerValues[0]
		}
	}
	return result
}
