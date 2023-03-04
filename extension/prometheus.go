package extension

import (
	"sync"
	"time"

	"github.com/flowscan/lapis"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	Success  = "success"
	NotFound = "notfound"
	Error    = "error"
)

// Collection of prometheus metrics
type StoreMetrics struct {
	AdditionalLabels        []string
	LoadTimeHistogram       *prometheus.HistogramVec
	LoadBatchHistogram      *prometheus.HistogramVec
	SetTimeHistogram        *prometheus.HistogramVec
	SetBatchHistogram       *prometheus.HistogramVec
	LayerLoadTimeHistogram  *prometheus.HistogramVec
	LayerLoadBatchHistogram *prometheus.HistogramVec
	LayerSetTimeHistogram   *prometheus.HistogramVec
	LayerSetBatchHistogram  *prometheus.HistogramVec
}

// Create a new store metric collector
// additionalLabels is a list of additional labels used for metric partitioning
func NewStoreMetrics(additionalLabels ...string) *StoreMetrics {
	c := &StoreMetrics{}
	c.AdditionalLabels = additionalLabels
	c.LoadTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Name:      "load_time_seconds",
		Help:      "The time it takes to resolve a load request",
	}, append(additionalLabels, []string{"store", "status"}...))
	c.LoadBatchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Name:      "load_batch",
		Help:      "The batch size for each load",
	}, append(additionalLabels, []string{"store"}...))
	c.LayerLoadTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Subsystem: "layer",
		Name:      "load_time_seconds",
		Help:      "The time a layer takes to resolve a load request",
	}, append(additionalLabels, []string{"store", "layer", "status"}...))
	c.LayerLoadBatchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Subsystem: "layer",
		Name:      "load_batch",
		Help:      "The batch size for each load on to a layer",
	}, append(additionalLabels, []string{"store", "layer"}...))
	c.SetTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Name:      "set_time_seconds",
		Help:      "The time it takes to resolve a set request",
	}, append(additionalLabels, []string{"store", "status"}...))
	c.SetBatchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Name:      "set_batch",
		Help:      "The batch size for each set",
	}, append(additionalLabels, []string{"store"}...))
	c.LayerSetTimeHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Subsystem: "layer",
		Name:      "set_time_seconds",
		Help:      "The time a layer takes to resolve a set request",
	}, append(additionalLabels, []string{"store", "layer", "status"}...))
	c.LayerSetBatchHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lapis",
		Subsystem: "layer",
		Name:      "set_batch",
		Help:      "The batch size for each set on to a layer",
	}, append(additionalLabels, []string{"store", "layer"}...))
	return c
}

// PrometheusMetrics is an extension for layer instrumentation
type PrometheusMetrics[TKey comparable, TValue any] struct {
	storeName          string
	metrics            *StoreMetrics
	labelValues        []string
	layerIdentifiers   []string
	layerLoadStartTime []map[uint64]time.Time
	layerLoadMu        []sync.Mutex
	layerSetStartTime  []map[uint64]time.Time
	layerSetMu         []sync.Mutex
	loadStartTime      map[uint64]time.Time
	setStartTime       map[uint64]time.Time
	loadMu             sync.Mutex
	setMu              sync.Mutex
}

// Create a new prometheus metrics extension with the given metrics collector
// labels is an optional parameter to add the given labels into the metrics from this store
func NewPrometheusMetrics[TKey comparable, TValue any](metrics *StoreMetrics, labelValues ...string) *PrometheusMetrics[TKey, TValue] {
	return &PrometheusMetrics[TKey, TValue]{
		labelValues: labelValues,
		metrics:     metrics,
	}
}

func (e *PrometheusMetrics[TKey, TValue]) Name() string { return "PrometheusMetrics" }

func (e *PrometheusMetrics[TKey, TValue]) InitializationHook(r *lapis.Store[TKey, TValue], layers []lapis.Layer[TKey, TValue]) error {
	e.storeName = r.Identifier()
	e.layerLoadStartTime = make([]map[uint64]time.Time, len(layers))
	e.layerSetStartTime = make([]map[uint64]time.Time, len(layers))
	e.layerLoadMu = make([]sync.Mutex, len(layers))
	e.layerSetMu = make([]sync.Mutex, len(layers))
	e.layerIdentifiers = make([]string, len(layers))
	e.loadStartTime = make(map[uint64]time.Time)
	e.setStartTime = make(map[uint64]time.Time)
	for i, layer := range layers {
		e.layerIdentifiers[i] = layer.Identifier()
		e.layerLoadStartTime[i] = make(map[uint64]time.Time)
		e.layerSetStartTime[i] = make(map[uint64]time.Time)
	}
	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) PreLoadHook(traceID uint64, keys []TKey) []error {
	// record the batch size
	e.metrics.LoadBatchHistogram.WithLabelValues(append(e.labelValues, e.storeName)...).Observe(float64(len(keys)))

	// record the start time for this trace
	e.loadMu.Lock()
	e.loadStartTime[traceID] = time.Now()
	e.loadMu.Unlock()
	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) PostLoadHook(traceID uint64, keys []TKey, values []TValue, errors []error) []error {
	// record the duration of the trace
	e.loadMu.Lock()
	traceTime := time.Since(e.loadStartTime[traceID]).Seconds()
	delete(e.loadStartTime, traceID)
	e.loadMu.Unlock()

	// record the status and trace for each key
	var status string
	for i := 0; i < len(keys); i++ {
		if len(errors) == 0 || (errors[i] == nil) {
			status = Success
		} else if _, ok := errors[i].(lapis.ErrNotFound[TKey]); ok {
			status = NotFound
		} else {
			status = Error
		}
		e.metrics.LoadTimeHistogram.WithLabelValues(append(e.labelValues, e.storeName, status)...).Observe(traceTime)
	}

	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) PreSetHook(traceID uint64, keys []TKey, values []TValue) []error {
	// record the batch size
	e.metrics.SetBatchHistogram.WithLabelValues(append(e.labelValues, e.storeName)...).Observe(float64(len(keys)))

	// record the start time for this trace
	e.setMu.Lock()
	e.setStartTime[traceID] = time.Now()
	e.setMu.Unlock()
	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) PostSetHook(traceID uint64, keys []TKey, values []TValue, errors []error) {
	// record the duration of the trace
	e.setMu.Lock()
	traceTime := time.Since(e.setStartTime[traceID]).Seconds()
	delete(e.setStartTime, traceID)
	e.setMu.Unlock()

	// record the status and trace for each key
	var status string
	for i := 0; i < len(keys); i++ {
		if len(errors) == 0 || (errors[i] == nil) {
			status = Success
		} else {
			status = Error
		}
		e.metrics.LayerSetTimeHistogram.WithLabelValues(append(e.labelValues, e.storeName, status)...).Observe(traceTime)
	}
}

func (e *PrometheusMetrics[TKey, TValue]) LayerPreLoadHook(traceID uint64, layerIndex int, keys []TKey) []error {
	// record the batch size
	e.metrics.LayerLoadBatchHistogram.WithLabelValues(append(e.labelValues, e.storeName, e.layerIdentifiers[layerIndex])...).Observe(float64(len(keys)))

	// record the start time for this trace
	e.layerLoadMu[layerIndex].Lock()
	e.layerLoadStartTime[layerIndex][traceID] = time.Now()
	e.layerLoadMu[layerIndex].Unlock()
	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) LayerPostLoadHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) []error {
	// record the duration of the trace
	e.layerLoadMu[layerIndex].Lock()
	traceTime := time.Since(e.layerLoadStartTime[layerIndex][traceID]).Seconds()
	delete(e.layerLoadStartTime[layerIndex], traceID)
	e.layerLoadMu[layerIndex].Unlock()

	// record the status and trace for each key
	var status string
	for i := 0; i < len(keys); i++ {
		if len(errors) == 0 || (errors[i] == nil) {
			status = Success
		} else if _, ok := errors[i].(lapis.ErrNotFound[TKey]); ok {
			status = NotFound
		} else {
			status = Error
		}
		e.metrics.LayerLoadTimeHistogram.WithLabelValues(append(e.labelValues, e.storeName, e.layerIdentifiers[layerIndex], status)...).Observe(traceTime)
	}

	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) LayerPreSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue) []error {
	// record the batch size
	e.metrics.LayerSetBatchHistogram.WithLabelValues(append(e.labelValues, e.storeName, e.layerIdentifiers[layerIndex])...).Observe(float64(len(keys)))

	// record the start time for this trace
	e.layerSetMu[layerIndex].Lock()
	e.layerSetStartTime[layerIndex][traceID] = time.Now()
	e.layerSetMu[layerIndex].Unlock()
	return nil
}

func (e *PrometheusMetrics[TKey, TValue]) LayerPostSetHook(traceID uint64, layerIndex int, keys []TKey, values []TValue, errors []error) {
	// record the duration of the trace
	e.layerSetMu[layerIndex].Lock()
	traceTime := time.Since(e.layerLoadStartTime[layerIndex][traceID]).Seconds()
	delete(e.layerSetStartTime[layerIndex], traceID)
	e.layerSetMu[layerIndex].Unlock()

	// record the status and trace for each key
	var status string
	for i := 0; i < len(keys); i++ {
		if len(errors) == 0 || (errors[i] == nil) {
			status = Success
		} else {
			status = Error
		}
		e.metrics.LayerSetTimeHistogram.WithLabelValues(append(e.labelValues, e.storeName, e.layerIdentifiers[layerIndex], status)...).Observe(traceTime)
	}
}
