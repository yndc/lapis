package lapis_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	squareMockRepository := newSquareMockRepository(t, 100*time.Millisecond)
	_, err := squareMockRepository.LoadAll([]int{1, 3, 5})
	_, err = squareMockRepository.LoadAll([]int{1, 2, 3, 4, 5})
	_, err = squareMockRepository.LoadAll([]int{2, 6, 4, 8, 9})
	_, err = squareMockRepository.LoadAll([]int{2, 6, 4, 8, 9})
	_, err = squareMockRepository.LoadAll([]int{6, 7, 8, 9, 0})
	r, err := squareMockRepository.LoadAll([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	assert.Equal(t, err, []error{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil})
	assert.Equal(t, r, []int{0, 1, 4, 9, 16, 25, 36, 49, 64, 81})
}

func TestBatch(t *testing.T) {
	squareMockRepository := newSquareMockRepository(t, 100*time.Millisecond)
	m := 10
	n := 50000

	for i := 0; i < m; i++ {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for j := 0; j < n; j++ {
			capturedIndex := j
			go func() {
				res, err := squareMockRepository.Load(capturedIndex)
				assert.Nil(t, err)
				assert.Equal(t, capturedIndex*capturedIndex, res)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func TestToSingleBatch(t *testing.T) {
	squareMockRepository := newSquareMockRepository(t, 1*time.Second)
	n := 100
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			time.Sleep(randDuration(0, 800*time.Millisecond))
			res, err := squareMockRepository.Load(10)
			assert.Nil(t, err)
			assert.Equal(t, 100, res)
			wg.Done()
		}()
	}
	wg.Wait()
}

func randDuration(from time.Duration, to time.Duration) time.Duration {
	delta := to - from
	return from + time.Duration(rand.Float64()*float64(delta))
}
