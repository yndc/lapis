package lapis

import "sync"

// Set a set of data to all of layers
// Returns an array of array of errors with the first dimension as the key and second dimension as the layer
func (r *Repository[TKey, TValue]) SetAll(keys []TKey, values []TValue, options ...SetOption) [][]error {
	layerIndexes := make([]int, len(r.layers))
	if sequential := GetOption[Sequential](options, SetOptionSequential); sequential != nil {
		var initial, iterationChange, endCondition int
		if ascending := GetOption[Ascending](options, SetOptionAscending); ascending != nil {
			initial = 0
			iterationChange = 1
			endCondition = len(r.layers)
		} else {
			initial = len(r.layers) - 1
			iterationChange = -1
			endCondition = -1
		}
		i := 0
		for layerIndex := initial; layerIndex != endCondition; layerIndex = layerIndex + iterationChange {
			layerIndexes[i] = layerIndex
			i++
		}
		return r.set(layerIndexes, keys, values, true)
	} else {
		layerIndexes = generateSequence(len(r.layers))
		return r.set(layerIndexes, keys, values, false)
	}
}

func (r *Repository[TKey, TValue]) set(layerIndexes []int, keys []TKey, values []TValue, sequential bool) [][]error {
	var traceID uint64 = r.getTraceID()
	var errors = make([][]error, len(r.layers))

	// execute pre-set hook
	if len(r.preSetHooks) > 0 {
		// TODO block execution for error-returning
		for _, hook := range r.preSetHooks {
			hook.PreSetHook(traceID, keys, values)
		}
	}

	if sequential {
		for i, layerIndex := range layerIndexes {
			errors[i] = r.layerSet(traceID, layerIndex, keys, values)
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(len(r.layers))
		for i, layerIndex := range layerIndexes {
			capturedI := i
			capturedLayerIndex := layerIndex
			go func() {
				defer wg.Done()
				errors[capturedI] = r.layerSet(traceID, capturedLayerIndex, keys, values)
			}()
		}
		wg.Wait()
	}

	// execute post-set hook
	if len(r.postSetHooks) > 0 {
		for _, hook := range r.postSetHooks {
			hook.PostSetHook(traceID, keys, values, errors)
		}
	}

	return errors
}

func (r *Repository[TKey, TValue]) layerSet(traceID uint64, layerIndex int, keys []TKey, values []TValue) []error {
	layer := r.layers[layerIndex]

	// execute layer pre-set hook
	if len(r.layerPreSetHooks) > 0 {
		// TODO block execution for error-returning
		for _, hook := range r.layerPreSetHooks {
			hook.LayerPreSetHook(traceID, layer, keys, values)
		}
	}

	// execute the layer set operation
	errors := layer.Set(keys, values)

	// execute layer post-set hook
	if len(r.layerPostSetHooks) > 0 {
		for _, hook := range r.layerPostSetHooks {
			hook.LayerPostSetHook(traceID, layer, keys, values, errors)
		}
	}

	return errors
}
