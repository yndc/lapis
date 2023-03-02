package extension

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger is an extension for lapis that logs access and value sets for debugging
type Logger[TKey comparable, TValue any] struct {
	layers                 []lapis.Layer[TKey, TValue]
	logger                 zerolog.Logger
	loggerMu               sync.RWMutex
	loggerLayerLoadStartAt []map[uint64]time.Time
	loggerLayerSetStartAt  []map[uint64]time.Time
}

func (e *Logger[TKey, TValue]) Name() string { return "Logger" }

func (e *Logger[TKey, TValue]) InitializationHook(r *lapis.Store[TKey, TValue], layers []lapis.Layer[TKey, TValue]) error {
	e.layers = layers
	e.loggerLayerLoadStartAt = make([]map[uint64]time.Time, len(layers))
	e.loggerLayerSetStartAt = make([]map[uint64]time.Time, len(layers))
	e.logger = log.With().Str("store", r.Identifier()).Logger()
	e.logger.Debug().Msgf("store initialized")
	for layerIndex := range layers {
		e.loggerLayerLoadStartAt[layerIndex] = make(map[uint64]time.Time)
		e.loggerLayerSetStartAt[layerIndex] = make(map[uint64]time.Time)
	}
	return nil
}

func (e *Logger[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	e.logger.Debug().Uint64("trace", traceID).Msgf("loading start: %v", keys)
	return nil
}

func (e *Logger[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	e.logger.Debug().Uint64("trace", traceID).Msgf("loading finish: %v (errors: %v)", values, errors)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPreLoadHook(traceID uint64, layerIndex int, keys []TKey) []error {
	e.loggerMu.Lock()
	e.loggerLayerLoadStartAt[layerIndex][traceID] = time.Now()
	e.loggerMu.Unlock()
	e.logger.Debug().Uint64("trace", traceID).Msgf("loading start at layer %v: %v", e.layers[layerIndex].Identifier(), keys)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPostLoadHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) []error {
	e.loggerMu.Lock()
	e.logger.Debug().Uint64("trace", traceID).Msgf(
		"loading finish from layer %v: %v (errors: %v) time: %v",
		e.layers[layerIndex].Identifier(),
		values,
		errors,
		time.Since(e.loggerLayerLoadStartAt[layerIndex][traceID]).Milliseconds(),
	)
	delete(e.loggerLayerLoadStartAt[layerIndex], traceID)
	e.loggerMu.Unlock()

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
	// 	// e.metrics.LayerLoadCounter.WithLabelValues(e.storeName, e.layerIdentifiers[layerIndex], status).Inc()
	// 	// e.metrics.LayerLoadTimeGauge.WithLabelValues(e.storeName, e.layerIdentifiers[layerIndex], status).Set(traceTime)
	// }
	return nil
}

func (e *Logger[TKey, TValue]) LayerPreSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue) []error {
	e.loggerMu.Lock()
	e.loggerLayerSetStartAt[layerIndex][traceID] = time.Now()
	e.loggerMu.Unlock()
	e.logger.Debug().Uint64("trace", traceID).Msgf("setting start at layer %v: keys: %v values: %v", e.layers[layerIndex].Identifier(), keys, values)
	return nil
}

func (e *Logger[TKey, TValue]) LayerPostSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) {
	e.loggerMu.Lock()
	e.logger.Debug().Uint64("trace", traceID).Msgf(
		"setting finish at layer %v: keys: %v values: %v errors: %v time: %v",
		e.layers[layerIndex].Identifier(),
		keys,
		values,
		errors,
		time.Since(e.loggerLayerSetStartAt[layerIndex][traceID]).Milliseconds(),
	)
	delete(e.loggerLayerSetStartAt[layerIndex], traceID)
	e.loggerMu.Unlock()
}
