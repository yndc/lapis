package tlru

// Wrapper for elements inside a priority queue, every data is paired with a priority value
type Item[T any] struct {
	priority int
	index    int
	Value    T
}

// Make a new priority queue item from the given value and priority
func NewItem[T any](value T, priority int) Item[T] {
	return Item[T]{
		priority: priority,
		Value:    value,
	}
}

// Check if the value inside the priority queue item is empty or invalid
func (i Item[T]) Empty() bool {
	return i.index == -1
}

// Priority queue of type T implementation
type PriorityQueue[T any] queue[T]

// Get the number of items in the priority queue
func (q PriorityQueue[T]) Len() int {
	return queue[T](q).Len()
}

// Push an entry to the queue
func (q *PriorityQueue[T]) Push(item Item[T]) {
	arr := queue[T](*q)
	pushHeap[Item[T]](&arr, item)
	*q = PriorityQueue[T](arr)
}

// Pop get the last entry in the queue
func (q *PriorityQueue[T]) Pop() Item[T] {
	arr := queue[T](*q)
	v := popHeap[Item[T]](&arr)
	*q = PriorityQueue[T](arr)
	return v
}

// Peek get the last entry in the queue without removing it from the collection
func (q *PriorityQueue[T]) Peek() Item[T] {
	return queue[T](*q).Peek()
}

// Create a new priority queue of type T
func NewQueue[T any]() *PriorityQueue[T] {
	q := make(queue[T], 0)
	initHeap[Item[T]](&q)
	pq := PriorityQueue[T](q)
	return &pq
}

type queue[T any] []Item[T]

func (q queue[T]) Len() int {
	return len(q)
}

func (q queue[T]) Less(i, j int) bool {
	return q[i].priority < q[j].priority
}

func (q queue[T]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue[T]) Push(item Item[T]) {
	n := len(*q)
	item.index = n
	*q = append(*q, item)
}

func (q *queue[T]) Pop() Item[T] {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1].index = -1
	item.index = -1
	*q = old[:n-1]
	return item
}

func (q queue[T]) Peek() Item[T] {
	if len(q) > 0 {
		return (q)[0]
	}
	return Item[T]{
		index: -1,
	}
}
