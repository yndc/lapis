package lapis

import "sync"

// Set a set of data to the layers, returns an array of array of errors with the first dimension as the key and second dimension as the layer
func (r *Repository[TKey, TValue]) SetAll(keys []TKey, values []TValue, options ...SetOption) [][]error {
	errors := make([][]error, len(r.layers))
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
		for i := initial; i != endCondition; i = i + iterationChange {
			layerErrors := r.layers[i].Set(keys, values)
			errors[i] = layerErrors
		}
	} else {
		wg := sync.WaitGroup{}
		wg.Add(len(r.layers))
		for index, layer := range r.layers {
			capturedIndex := index
			capturedLayer := layer
			go func() {
				layerErrors := capturedLayer.Set(keys, values)
				errors[capturedIndex] = layerErrors
			}()
		}
	}

	return errors
}
