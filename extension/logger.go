package extension

import (
	"github.com/flowscan/lapis"
	"github.com/rs/zerolog/log"
)

// Logger is an extension for lapis that logs access and value sets for debugging
type Logger[TKey comparable, TValue any] struct{}

func (e Logger[TKey, TValue]) Name() string { return "Logger" }

func (e Logger[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	log.Debug().Uint64("trace", traceID).Msgf("start loading: %v", keys)
	return nil
}

func (e Logger[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	log.Debug().Uint64("trace", traceID).Msgf("finish loading: %v (errors: %v)", values, errors)
	return nil
}

func (e Logger[TKey, TValue]) LayerPreLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey) []error {
	log.Debug().Uint64("trace", traceID).Msgf("start loading at layer %v: %v", layer.Identifier(), keys)
	return nil
}

func (e Logger[TKey, TValue]) LayerPostLoadHook(traceID uint64, layer lapis.Layer[TKey, TValue], keys []TKey, values []TValue, errors []error) []error {
	log.Debug().Uint64("trace", traceID).Msgf("finish loading from layer %v: %v (errors: %v)", layer.Identifier(), values, errors)
	return nil
}
