package extension

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/rs/zerolog/log"
)

var loggerLayerLoadStartAt map[string]map[uint64]time.Time
var loggerLayerSetStartAt map[string]map[uint64]time.Time
var loggerMu sync.RWMutex

// Logger is an extension for lapis that logs access and value sets for debugging
type Logger[TKey comparable, TValue any] struct {
	layers []lapis.Layer[TKey, TValue]
}

func (e *Logger[TKey, TValue]) Name() string { return "Logger" }

func (e *Logger[TKey, TValue]) InitializationHook(r *lapis.Repository[TKey, TValue], layers []lapis.Layer[TKey, TValue]) error {
	e.layers = layers
	loggerLayerLoadStartAt = make(map[string]map[uint64]time.Time)
	loggerLayerSetStartAt = make(map[string]map[uint64]time.Time)
	for _, layer := range layers {
		loggerLayerLoadStartAt[layer.Identifier()] = make(map[uint64]time.Time)
		loggerLayerSetStartAt[layer.Identifier()] = make(map[uint64]time.Time)
	}
	log.Debug().Msgf("repository initialized")
	return nil
}

func (e *Logger[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	log.Debug().Uint64("trace", traceID).Msgf("loading start: %v", keys)
	return nil
}

func (e *Logger[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	log.Debug().Uint64("trace", traceID).Msgf("loading finish: %v (errors: %v)", values, errors)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPreLoadHook(traceID uint64, layerIndex int, keys []TKey) []error {
	loggerMu.Lock()
	loggerLayerLoadStartAt[e.layers[layerIndex].Identifier()][traceID] = time.Now()
	loggerMu.Unlock()
	log.Debug().Uint64("trace", traceID).Msgf("loading start at layer %v: %v", e.layers[layerIndex].Identifier(), keys)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPostLoadHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) []error {
	loggerMu.Lock()
	log.Debug().Uint64("trace", traceID).Msgf(
		"loading finish from layer %v: %v (errors: %v) time: %v",
		e.layers[layerIndex].Identifier(),
		values,
		errors,
		time.Since(loggerLayerLoadStartAt[e.layers[layerIndex].Identifier()][traceID]).Milliseconds(),
	)
	delete(loggerLayerLoadStartAt[e.layers[layerIndex].Identifier()], traceID)
	loggerMu.Unlock()

	// // log the status for each result
	// var status string
	// for i := 0; i < len(keys); i++ {
	// 	if len(errors) == 0 || (errors[i] == nil) {
	// 		status = Success
	// 	} else if _, ok := errors[i].(lapis.ErrNotFound[TKey]); ok {
	// 		status = NotFound
	// 	} else {
	// 		status = Error
	// 	}
	// 	// e.metrics.LayerLoadCounter.WithLabelValues(e.repositoryName, e.layerIdentifiers[layerIndex], status).Inc()
	// 	// e.metrics.LayerLoadTimeGauge.WithLabelValues(e.repositoryName, e.layerIdentifiers[layerIndex], status).Set(traceTime)
	// }
	return nil
}

func (e *Logger[TKey, TValue]) LayerPreSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue) []error {
	loggerMu.Lock()
	loggerLayerSetStartAt[e.layers[layerIndex].Identifier()][traceID] = time.Now()
	loggerMu.Unlock()
	log.Debug().Uint64("trace", traceID).Msgf("setting start at layer %v: keys: %v values: %v", e.layers[layerIndex].Identifier(), keys, values)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPostSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) {
	loggerMu.Lock()
	log.Debug().Uint64("trace", traceID).Msgf(
		"setting finish at layer %v: keys: %v values: %v errors: %v time: %v",
		e.layers[layerIndex].Identifier(),
		keys,
		values,
		errors,
		time.Since(loggerLayerSetStartAt[e.layers[layerIndex].Identifier()][traceID]).Milliseconds(),
	)
	delete(loggerLayerSetStartAt[e.layers[layerIndex].Identifier()], traceID)
	loggerMu.Unlock()
}
