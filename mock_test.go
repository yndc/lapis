package lapis_test

import (
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/extension"
	"github.com/flowscan/lapis/layer"
)

// create a simple int -> int store with 2 layers, a in-memory cache and a backend that squares integers with 100ms delay
func newSquareMockStore(t *testing.T, fakeDelay time.Duration) *lapis.Store[int, int] {
	store, err := lapis.New(lapis.Config[int, int]{
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 10 * time.Hour}),
			SquareMockBackend{fakeDelay: fakeDelay},
		},
		Extensions: []lapis.Extension{
			&extension.Logger[int, int]{},
		},
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
		},
		Identifier: "SquareMockStore",
	})
	if err != nil {
		t.Fatalf("failed to create mock square store")
	}

	return store
}

type SquareMockBackend struct {
	fakeDelay time.Duration
}

func (s SquareMockBackend) Identifier() string { return "SquareMockBackend" }

func (s SquareMockBackend) Get(keys []int) ([]int, []error) {
	time.Sleep(s.fakeDelay)
	result := make([]int, len(keys))
	for i, v := range keys {
		result[i] = v * v
	}
	return result, nil
}

func (s SquareMockBackend) Set(keys []int, values []int) []error {
	return nil
}

type UnstableBackend struct {
	fakeDelay          time.Duration
	successProbability float64
}

func (s UnstableBackend) Identifier() string {
	return fmt.Sprintf("UnstableBackend %f", s.successProbability)
}

func (s UnstableBackend) Get(keys []int) ([]int, []error) {
	time.Sleep(s.fakeDelay)
	result := make([]int, len(keys))
	errors := make([]error, len(keys))
	for i, key := range keys {
		r := rand.Float64()
		if r < s.successProbability {
			result[i] = key * key
		} else {
			errors[i] = lapis.NewErrNotFound(key)
		}
	}
	return result, errors
}

func (s UnstableBackend) Set(keys []int, values []int) []error {
	return nil
}

type SettableBackend struct {
	fakeDelay  time.Duration
	multiplier int32
}

func (s *SettableBackend) Identifier() string {
	return "SettableBackend"
}

func (s *SettableBackend) Get(keys []int) ([]int, []error) {
	time.Sleep(s.fakeDelay)
	result := make([]int, len(keys))
	errors := make([]error, len(keys))
	for i, key := range keys {
		result[i] = int(s.multiplier) * key
	}
	return result, errors
}

func (s *SettableBackend) Set(keys []int, values []int) []error {
	return nil
}

func (s *SettableBackend) SetMultiplier(value int32) {
	atomic.AddInt32(&s.multiplier, value-s.multiplier)
}

type NotPrimeOnlyBackend struct {
	fakeDelay time.Duration
}

func (s *NotPrimeOnlyBackend) Identifier() string {
	return "IsNotPrimeOnlyBackend"
}

func (s *NotPrimeOnlyBackend) Get(keys []int) ([]int, []error) {
	time.Sleep(s.fakeDelay)
	result := make([]int, len(keys))
	errors := make([]error, len(keys))
	for i, key := range keys {
		if isPrime(key) {
			errors[i] = lapis.NewErrNotFound(key)
		} else {
			result[i] = key
		}
	}
	return result, errors
}

func (s *NotPrimeOnlyBackend) Set(keys []int, values []int) []error {
	return nil
}

func isPrime(x int) bool {
	if x < 4 {
		return true
	}
	sq_root := int(math.Sqrt(float64(x)))
	for i := 2; i <= sq_root; i++ {
		if x%i == 0 {
			return false
		}
	}
	return true
}
