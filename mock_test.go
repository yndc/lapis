package lapis_test

import (
	"testing"
	"time"

	"github.com/flowscan/lapis"
	"github.com/flowscan/lapis/layer"
)

// create a simple int -> int repository with 2 layers, a in-memory cache and a backend that squares integers with 100ms delay
func newSquareMockRepository(t *testing.T) *lapis.Repository[int, int] {
	repository, err := lapis.New(lapis.Config[int, int]{
		Layers: []lapis.Layer[int, int]{
			layer.NewMemory[int, int](layer.MemoryConfig{Retention: 10 * time.Hour}),
			SquareMockBackend{fakeDelay: 100 * time.Millisecond},
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
