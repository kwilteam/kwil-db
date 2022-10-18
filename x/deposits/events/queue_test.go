package events_test

import (
	"kwil/x/deposits/events"
	"testing"
)

func TestQueue(t *testing.T) {
	q := events.NewQueue()

	// Test append
	q.Append(1)
	q.Append(2)
	q.Append(3)

	if q.Len() != 3 {
		t.Errorf("expected length of 3, got %d", q.Len())
	}

	// testing head and tail
	if q.Head() != 1 {
		t.Errorf("expected head of 1, got %d", q.Head())
	}

	if q.Tail() != 3 {
		t.Errorf("expected tail of 3, got %d", q.Tail())
	}

	// Test pop
	if q.Pop() != 1 {
		t.Errorf("expected pop of 1, got %d", q.Pop())
	}

	if q.Pop() != 2 {
		t.Errorf("expected pop of 2, got %d", q.Pop())
	}

	if q.Pop() != 3 {
		t.Errorf("expected pop of 3, got %d", q.Pop())
	}

	if q.Pop() != -1 {
		t.Error("expected -1 when popping empty queue")
	}

}
