package lib 

import "testing"

func TestQueueEnqueue(t *testing.T){
    q := NewQueue[int]()

    q.Enqueue(10)
    q.Enqueue(10)
    q.Enqueue(10)
    
    if len(q.items) != 3 {
        t.Errorf("Expected queue length 3, got %d", len(q.items))
    }
}

func TestQueueDequeue(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	item, ok := q.Dequeue()
	if !ok || item != 1 {
		t.Errorf("Expected 1, got %v", item)
	}

	item, ok = q.Dequeue()
	if !ok || item != 2 {
		t.Errorf("Expected 2, got %v", item)
	}

	item, ok = q.Dequeue()
	if !ok || item != 3 {
		t.Errorf("Expected 3, got %v", item)
	}

	_, ok = q.Dequeue()
	if ok {
		t.Errorf("Expected Dequeue to return false on empty queue")
	}
}

func TestQueueIsEmpty(t *testing.T) {
	q := NewQueue[string]()

	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty")
	}

	q.Enqueue("hello")
	if q.IsEmpty() {
		t.Errorf("Expected queue to be non-empty")
	}

	q.Dequeue()
	if !q.IsEmpty() {
		t.Errorf("Expected queue to be empty after removing element")
	}
}
