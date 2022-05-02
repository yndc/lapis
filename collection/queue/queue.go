package queue

import (
	"sync"
)

type Queue[T any] struct {
	count int // Number of items inside this queue
	head  int // queue head index in the array
	tail  int // queue tail index in the array
	cap   int // capacity of the queue
	arr   []T
	mu    sync.RWMutex
}

// creates a new queue with the given capacity
func NewQueue[T any](cap int) *Queue[T] {
	cap = zeroFallback(cap, 1)
	b := &Queue[T]{}
	b.arr = make([]T, cap)
	b.cap = cap
	return b
}

// Get the number of items in the buffer
func (b *Queue[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}

// Write a new data into the queue
func (b *Queue[T]) Enqueue(data T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.arr[b.head] = data
	nextHead := wrap(b.head+1, b.cap)
	if nextHead == b.tail {
		b.grow()
	} else {
		b.head = nextHead
	}
	b.count++
}

// Retrieve data from the queue, undefined behaviour if the queue is empty
func (b *Queue[T]) Dequeue() T {
	b.mu.Lock()
	defer b.mu.Unlock()
	data := b.arr[b.tail]
	b.tail = wrap(b.tail+1, b.cap)
	b.count--
	return data
}

// Peek the earliest data in the queue without removing it from the queue, undefined behaviour if the queue is empty
func (b *Queue[T]) Peek() T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	data := b.arr[b.tail]
	return data
}

// Peek the data at the given index without removing it from the queue. Panics if the index is outside the queue boundary
func (b *Queue[T]) At(index int) T {
	b.mu.RLock()
	defer b.mu.RUnlock()
	data := b.arr[wrap(b.tail+index, b.cap)]
	return data
}

// grow the array that backs the queue
func (b *Queue[T]) grow() {
	oldCap := b.cap
	newCap := growCap(b.cap)
	newArr := make([]T, newCap)

	// copy all data from the previous queue to the new one
	cursor := b.tail
	for i := 0; i < b.cap; i++ {
		newArr[i] = b.arr[cursor]
		cursor = wrap(cursor+1, b.cap)
	}

	b.arr = newArr
	b.cap = newCap
	b.tail = 0
	b.head = oldCap
}

func growCap(prev int) int {
	if prev < 1024 {
		return 2 * prev
	}
	return int(float64(prev) * 1.25)
}

func wrap(n int, cap int) int {
	if n < 0 {
		return (cap - ((n * -1) % cap)) % cap
	}
	return n % cap
}

func zeroFallback[T comparable](v T, d T) T {
	var zeroValue T
	if v == zeroValue {
		return d
	}
	return v
}
