package collection

import (
	"testing"
)

// test function
func TestUnsafeQueue(t *testing.T) {
	queue := NewQueueUnsafe[int]()
	handleQueueCalls(t, queue)
}

// ToDo: add concurrency tests
func TestSafeQueue(t *testing.T) {
	queue := NewQueue[int]()
	handleQueueCalls(t, queue)
}

// test function
func handleQueueCalls(t *testing.T, queue Queue[int]) {
	if !queue.IsEmpty() {
		t.Errorf("Queue should be empty")
	}

	s, ok := queue.Poll()
	if ok {
		t.Errorf("Expected not 'ok', got %d", s)
	}

	queue.Add(10)
	queue.AddIf(20, func(v int) bool {
		return v <= 10
	})

	if queue.IsEmpty() {
		t.Errorf("Queue should NOT be empty")
	}

	if queue.Size() != 1 {
		t.Errorf("Queue size incorrect, expected %d, got %d", 1, queue.Size())
	}

	s, ok = queue.Poll()
	if !ok {
		t.Error("Expected 'ok'")
	}

	if s != 10 {
		t.Errorf("Expected %d, got %d", 10, s)
	}

	if !queue.IsEmpty() {
		t.Errorf("Queue should be empty")
	}

	s, ok = queue.Poll()
	if ok {
		t.Errorf("Expected not 'ok', got %d", s)
	}
}
