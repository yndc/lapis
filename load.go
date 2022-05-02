package lapis

import "context"

// Load a data from it's key
func (r *Repository[TKey, TValue]) Load(key TKey, options ...LoadOption) (TValue, error) {
	if UseNoBatch(options) {
		return Singlify(r.resolve)(key)
	}
	return r.batcher.Load(key)
}

// Load a data by key with a context, if the context is cancelled, the data loading will be cancelled too if
// this is the only operation
func (r *Repository[TKey, TValue]) LoadCtx(ctx context.Context, key TKey, options ...LoadOption) (TValue, error) {
	panic("not implemented")
}

// Load a set of data from their keys
func (r *Repository[TKey, TValue]) LoadAll(keys []TKey, options ...LoadOption) ([]TValue, []error) {
	if UseNoBatch(options) {
		return r.resolve(keys)
	}
	return r.batcher.LoadAll(keys)
}

// Load a set of data from their keys and prime the layers with the data resolved by the next layer
func (r *Repository[TKey, TValue]) resolve(keys []TKey) ([]TValue, []error) {
	var keysCount = len(keys)
	var result = make([]TValue, keysCount) // array containing the final result of values
	var errors = make([]error, keysCount)  // array containing errors for each keys

	var resultIndexes = generateSequence(len(keys)) // an array of indexes from the current layer's array to the original result array
	var layerKeys = keys                            // set of keys to be resolved by the current layer

	// iterate over all data layers from the beginning to the end
	// if any of the results are empty, try resolving the data from the next layer
	for layerIndex, layer := range r.layers {
		layerResult, layerErrors := layer.Get(layerKeys)
		resolvedLayerIndexes, resolvedLayerKeys, resolvedLayerValues, unresolvedLayerIndexes, unresolvedLayerKeys := group(layerKeys, layerResult, layerErrors)

		if len(resolvedLayerKeys) > 0 {
			resolvedResultIndexes := extract(resultIndexes, resolvedLayerIndexes)

			// merge the resolved values to the result
			merge(result, resolvedLayerValues, resolvedResultIndexes)

			// clear all errors from previous layers
			setZero(errors, resolvedResultIndexes)

			// prime the data on the previous layers
			if layerIndex > 0 {
				for i := layerIndex - 1; i >= 0; i-- {
					go r.layers[i].Set(resolvedLayerKeys, resolvedLayerValues)
				}
			}

			// skip going into the next layers if all data is already resolved
			if len(resolvedLayerKeys) == len(layerKeys) {
				break
			}
		}

		// load the unresolved data from the next layer
		layerKeys = unresolvedLayerKeys
		resultIndexes = extract(resultIndexes, unresolvedLayerIndexes)
	}

	return result, errors
}

// extract a layer resolver result
func group[TKey comparable, TValue any](keys []TKey, values []TValue, errors []error) (
	[]int,
	[]TKey,
	[]TValue,
	[]int,
	[]TKey,
) {
	resolvedIndexes := make([]int, len(keys))
	resolvedKeys := make([]TKey, len(keys))
	resolvedValues := make([]TValue, len(keys))
	resolvedCounter := 0
	unresolvedIndexes := make([]int, len(keys))
	unresolvedKeys := make([]TKey, len(keys))
	unresolvedCounter := 0
	for i := range keys {
		if len(errors) == 0 || (errors[i] == nil) {
			resolvedIndexes[resolvedCounter] = i
			resolvedKeys[resolvedCounter] = keys[i]
			resolvedValues[resolvedCounter] = values[i]
			resolvedCounter++
		} else {
			unresolvedIndexes[unresolvedCounter] = i
			unresolvedKeys[unresolvedCounter] = keys[i]
			unresolvedCounter++
		}
	}
	return resolvedIndexes[:resolvedCounter],
		resolvedKeys[:resolvedCounter],
		resolvedValues[:resolvedCounter],
		unresolvedIndexes[:resolvedCounter],
		unresolvedKeys[:resolvedCounter]
}

// write the values from the source array into the destination array based on the given indexes
func merge[T any](destination []T, source []T, indexes []int) {
	for i, dstIndex := range indexes {
		destination[dstIndex] = source[i]
	}
}

// set the elements in the destination array to setZero
func setZero[T any](destination []T, indexes []int) {
	var zero T
	for _, dstIndex := range indexes {
		destination[dstIndex] = zero
	}
}

// extract an array from the original array using the given indexes
func extract[T any](source []T, indexes []int) []T {
	result := make([]T, len(indexes))
	for i, v := range indexes {
		result[i] = source[v]
	}
	return result
}

// generate a sequence of integers starting from 0
func generateSequence(count int) []int {
	arr := make([]int, count)
	for i := 0; i < count; i++ {
		arr[i] = i
	}
	return arr
}

// Wait for all of the given contexes to be closed
func WaitAll(chans ...<-chan struct{}) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		for _, c := range chans {
			<-c
		}
		close(done)
	}()

	return done
}
