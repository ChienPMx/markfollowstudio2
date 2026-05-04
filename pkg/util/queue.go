package util

type Queue[T any] interface {
	Enqueue(item T) bool
	Dequeue() (T, bool)
	IsEmpty() bool
	IsFull() bool
	Size() int
	Peek() (T, bool) // Optional, add as needed
}

type CircularQueue[T any] struct {
	data  []T
	front int // Front pointer
	rear  int // Rear pointer
	count int // Tracks element count
}

// NewCircularQueue creates a new circular queue
func NewCircularQueue[T any](maxSize int) *CircularQueue[T] {
	return &CircularQueue[T]{
		data:  make([]T, maxSize),
		front: 0,
		rear:  0,
		count: 0,
	}
}

func (q *CircularQueue[T]) IsEmpty() bool { return q.count == 0 }
func (q *CircularQueue[T]) IsFull() bool  { return q.count == len(q.data) }
func (q *CircularQueue[T]) Size() int     { return q.count }

func (q *CircularQueue[T]) Enqueue(item T) bool {
	if q.IsFull() {
		return false
	}
	q.data[q.rear] = item
	q.rear = (q.rear + 1) % len(q.data)
	q.count++
	return true
}

func (q *CircularQueue[T]) Dequeue() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}
	item := q.data[q.front]
	q.front = (q.front + 1) % len(q.data)
	q.count--
	return item, true
}

func (q *CircularQueue[T]) Peek() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}
	return q.data[q.front], true
}
