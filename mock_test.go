package lapis_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/extension"
	"github.com/flowscan/lapis/layer"
)

// create a simple int -> int repository with 2 layers, a in-memory cache and a backend that squares integers with 100ms delay
func newSquareMockRepository(t *testing.T, fakeDelay time.Duration) *lapis.Repository[int, int] {
	repository, err := lapis.New(lapis.Config[int, int]{
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 10 * time.Hour}),
			SquareMockBackend{fakeDelay: fakeDelay},
		},
		Extensions: []lapis.Extension{
			extension.Logger[int, int]{},
		},
		Batcher: &lapis.BatcherConfig[int, int]{
			MaxBatch: 256,
		},
	})
	if err != nil {
		t.Fatalf("failed to create mock square repository")
	}

	return repository
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
