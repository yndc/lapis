package lapis_test

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/extension"
	"github.com/flowscan/lapis/layer"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixMicro())
	code := m.Run()
	os.Exit(code)
}

func TestLoad(t *testing.T) {
	squareMockStore := newSquareMockStore(t, 100*time.Millisecond)
	squareMockStore.LoadAll([]int{1, 3, 5})
	squareMockStore.LoadAll([]int{1, 2, 3, 4, 5})
	squareMockStore.LoadAll([]int{2, 6, 4, 8, 9})
	squareMockStore.LoadAll([]int{2, 6, 4, 8, 9})
	squareMockStore.LoadAll([]int{6, 7, 8, 9, 0})
	r, err := squareMockStore.LoadAll([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	assert.Equal(t, err, []error{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil})
	assert.Equal(t, r, []int{0, 1, 4, 9, 16, 25, 36, 49, 64, 81})
}

func TestBatch(t *testing.T) {
	squareMockStore := newSquareMockStore(t, 100*time.Millisecond)
	m := 10
	n := 50000

	for i := 0; i < m; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				res, err := squareMockStore.Load(capturedIndex)
				assert.Nil(t, err)
				assert.Equal(t, capturedIndex*capturedIndex, res)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func TestToSingleBatch(t *testing.T) {
	squareMockStore := newSquareMockStore(t, 1000*time.Millisecond)
	n := 10000
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			time.Sleep(randDuration(0, 800*time.Millisecond))
			res, err := squareMockStore.Load(10)
			assert.Nil(t, err)
			assert.Equal(t, 100, res)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestDeepLayers(t *testing.T) {
	store, err := lapis.New(lapis.Config[int, int]{
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
		},
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 100 * time.Millisecond}),
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.1},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.2},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.3},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.4},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.5},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.6},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.7},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.8},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 0.9},
			UnstableBackend{fakeDelay: 100 * time.Millisecond, successProbability: 1.0},
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
		},
		Identifier: "DeepLayerStore",
	})

	assert.Nil(t, err)

	m := 10
	n := 100
	for i := 0; i < m; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				res, err := store.Load(capturedIndex)
				assert.Nil(t, err)
				assert.Equal(t, capturedIndex*capturedIndex, res)
				wg.Done()
			}()
		}
		wg.Wait()
	}

	time.Sleep(1000 * time.Millisecond)
}

func TestSet(t *testing.T) {
	backend := &SettableBackend{fakeDelay: 0 * time.Millisecond, multiplier: 1}
	store, err := lapis.New(lapis.Config[int, int]{
		Identifier: "TestSet",
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
			Wait:     10 * time.Millisecond,
		},
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 50 * time.Millisecond}),
			backend,
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
		},
	})

	assert.Nil(t, err)

	// e := 10
	m := 10
	n := 1000
	var errorCount int32 = 0
	var epoch int32 = 1
	var done chan struct{} = make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(n * m)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-done:
				return
			default:
				newValue := atomic.AddInt32(&epoch, 1)
				backend.SetMultiplier(newValue)
			}
		}
	}()
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				time.Sleep(randDuration(0, 3000*time.Millisecond))
				res, err := store.Load(capturedIndex)
				assert.Nil(t, err)
				if res != capturedIndex*int(epoch) {
					fmt.Printf("key: %v expected: %v (epoch %v, actual: %v)\n", capturedIndex, capturedIndex*int(epoch), epoch, res/capturedIndex)
					atomic.AddInt32(&errorCount, 1)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
	close(done)

	errorRatePct := int(errorCount) * 100 / (m * n)
	assert.True(t, errorRatePct < 10)
}

func TestPartialError(t *testing.T) {
	backend := &NotPrimeOnlyBackend{fakeDelay: 100 * time.Millisecond}
	store, err := lapis.New(lapis.Config[int, int]{
		Identifier: "PartialError",
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
			Wait:     10 * time.Millisecond,
		},
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 50 * time.Millisecond}),
			backend,
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
		},
	})

	assert.Nil(t, err)

	m := 10
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n * m)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				time.Sleep(randDuration(0, 1000*time.Millisecond))
				res, err := store.Load(capturedIndex)
				if isPrime(capturedIndex) {
					assert.NotNil(t, err)
					assert.Equal(t, res, 0)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, res, capturedIndex)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func TestPrometheus(t *testing.T) {
	backend := &NotPrimeOnlyBackend{fakeDelay: 100 * time.Millisecond}
	metrics := extension.NewStoreMetrics()
	store, err := lapis.New(lapis.Config[int, int]{
		Identifier: "TestPrometheus",
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
			Wait:     10 * time.Millisecond,
		},
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 50 * time.Millisecond}),
			backend,
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
			extension.NewPrometheusMetrics[int, int](metrics),
		},
	})

	assert.Nil(t, err)

	m := 10
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n * m)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				time.Sleep(randDuration(0, 1000*time.Millisecond))
				res, err := store.Load(capturedIndex)
				if isPrime(capturedIndex) {
					assert.NotNil(t, err)
					assert.Equal(t, res, 0)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, res, capturedIndex)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()

	c := make(chan prometheus.Metric)
	go metrics.LayerLoadBatchHistogram.Collect(c)
	go metrics.LayerLoadTimeHistogram.Collect(c)
	go func() {
		for {
			printMetric(<-c)
		}
	}()

	time.Sleep(1000 * time.Millisecond)
}

func TestPrometheusWithLabels(t *testing.T) {
	backend := &NotPrimeOnlyBackend{fakeDelay: 100 * time.Millisecond}
	metrics := extension.NewStoreMetrics("one", "two")
	store, err := lapis.New(lapis.Config[int, int]{
		Identifier: "TestPrometheus",
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
			Wait:     10 * time.Millisecond,
		},
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 50 * time.Millisecond}),
			backend,
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
			extension.NewPrometheusMetrics[int, int](metrics, "1", "2"),
		},
	})

	assert.Nil(t, err)

	m := 10
	n := 1000
	wg := sync.WaitGroup{}
	wg.Add(n * m)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				time.Sleep(randDuration(0, 1000*time.Millisecond))
				res, err := store.Load(capturedIndex)
				if isPrime(capturedIndex) {
					assert.NotNil(t, err)
					assert.Equal(t, res, 0)
				} else {
					assert.Nil(t, err)
					assert.Equal(t, res, capturedIndex)
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()

	c := make(chan prometheus.Metric)
	go metrics.LayerLoadBatchHistogram.Collect(c)
	go metrics.LayerLoadTimeHistogram.Collect(c)
	go func() {
		for {
			printMetric(<-c)
		}
	}()

	time.Sleep(1000 * time.Millisecond)
}

func printMetric(metric prometheus.Metric) {
	val := &io_prometheus_client.Metric{}
	metric.Write(val)
	fmt.Printf("%v: %v\n", metric.Desc(), val)
}

func randDuration(from time.Duration, to time.Duration) time.Duration {
	delta := to - from
	return from + time.Duration(rand.Float64()*float64(delta))
}
