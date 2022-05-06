package extension

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/rs/zerolog/log"
)

var layerStartAt map[string]map[uint64]time.Time
var mu sync.RWMutex

// Logger is an extension for lapis that logs access and value sets for debugging
type Logger[TKey comparable, TValue any] struct{}

func (e Logger[TKey, TValue]) Name() string { return "Logger" }

func (e Logger[TKey, TValue]) InitializationHook(layers []lapis.Layer[TKey, TValue]) error {
	layerStartAt = make(map[string]map[uint64]time.Time)
	for _, layer := range layers {
		layerStartAt[layer.Identifier()] = make(map[uint64]time.Time)
	}
	return nil
}

func (e Logger[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	log.Debug().Uint64("trace", traceID).Msgf("start loading: %v", keys)
	return nil
}

func (e Logger[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	log.Debug().Uint64("trace", traceID).Msgf("finish loading: %v (errors: %v)", values, errors)
	return nil
}

func (e Logger[TKey, TValue]) LayerPreLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey) []error {
	mu.Lock()
	layerStartAt[layer.Identifier()][traceID] = time.Now()
	mu.Unlock()
	log.Debug().Uint64("trace", traceID).Msgf("start loading at layer %v", layer.Identifier())
	return nil
}

func (e Logger[TKey, TValue]) LayerPostLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) []error {
	mu.Lock()
	log.Debug().Uint64("trace", traceID).Msgf(
		"finish loading from layer %v: %v (errors: %v) time: %v",
		layer.Identifier(),
		values,
		errors,
		time.Since(layerStartAt[layer.Identifier()][traceID]).Milliseconds(),
	)
	delete(layerStartAt[layer.Identifier()], traceID)
	mu.Unlock()
	return nil
}
