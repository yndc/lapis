package extension

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/rs/zerolog/log"
)

var layerLoadStartAt map[string]map[uint64]time.Time
var layerSetStartAt map[string]map[uint64]time.Time
var mu sync.RWMutex

// Logger is an extension for lapis that logs access and value sets for debugging
type Logger[TKey comparable, TValue any] struct{}

func (e Logger[TKey, TValue]) Name() string { return "Logger" }

func (e Logger[TKey, TValue]) InitializationHook(r *lapis.Repository[TKey, TValue], layers []lapis.Layer[TKey, TValue]) error {
	layerLoadStartAt = make(map[string]map[uint64]time.Time)
	layerSetStartAt = make(map[string]map[uint64]time.Time)
	for _, layer := range layers {
		layerLoadStartAt[layer.Identifier()] = make(map[uint64]time.Time)
		layerSetStartAt[layer.Identifier()] = make(map[uint64]time.Time)
	}
	log.Debug().Msgf("repository initialized")
	return nil
}

func (e Logger[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	log.Debug().Uint64("trace", traceID).Msgf("loading start: %v", keys)
	return nil
}

func (e Logger[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	log.Debug().Uint64("trace", traceID).Msgf("loading finish: %v (errors: %v)", values, errors)
	return nil
}

func (e Logger[TKey, TValue]) LayerPreLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey) []error {
	mu.Lock()
	layerLoadStartAt[layer.Identifier()][traceID] = time.Now()
	mu.Unlock()
	log.Debug().Uint64("trace", traceID).Msgf("loading start at layer %v: %v", layer.Identifier(), keys)
	return nil
}

func (e Logger[TKey, TValue]) LayerPostLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) []error {
	mu.Lock()
	log.Debug().Uint64("trace", traceID).Msgf(
		"loading finish from layer %v: %v (errors: %v) time: %v",
		layer.Identifier(),
		values,
		errors,
		time.Since(layerLoadStartAt[layer.Identifier()][traceID]).Milliseconds(),
	)
	delete(layerLoadStartAt[layer.Identifier()], traceID)
	mu.Unlock()
	return nil
}

func (e Logger[TKey, TValue]) LayerPreSetHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey, values []TValue) []error {
	mu.Lock()
	layerSetStartAt[layer.Identifier()][traceID] = time.Now()
	mu.Unlock()
	log.Debug().Uint64("trace", traceID).Msgf("setting start at layer %v: keys: %v values: %v", layer.Identifier(), keys, values)
	return nil
}

func (e Logger[TKey, TValue]) LayerPostSetHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) {
	mu.Lock()
	log.Debug().Uint64("trace", traceID).Msgf(
		"setting finish at layer %v: keys: %v values: %v errors: %v time: %v",
		layer.Identifier(),
		keys,
		values,
		errors,
		time.Since(layerSetStartAt[layer.Identifier()][traceID]).Milliseconds(),
	)
	delete(layerSetStartAt[layer.Identifier()], traceID)
	mu.Unlock()
}
