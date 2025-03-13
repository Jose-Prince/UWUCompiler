package lib

type Queue[T any] struct {
    items []T
}

// Creates empty Queue
func NewQueue[T any]() *Queue[T] {
    return &Queue[T]{items: []T{}}
}

// Add item at end of Queue
func (q *Queue[T]) Enqueue(item T) {
    q.items = append(q.items, item)
}

// Eliminates and returns the first element of Queue
func (q *Queue[T]) Dequeue() (T, bool) {
    if q.IsEmpty() {
        var zeroValue T
        return zeroValue, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

// Verifies if Queue is empty
func (q *Queue[T]) IsEmpty() bool {
    return len(q.items) == 0
}
