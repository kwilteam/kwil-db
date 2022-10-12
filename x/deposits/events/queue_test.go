package events_test

import (
	"kwil/x/chain/events"
	"math/big"
	"testing"
)

func TestQueue(t *testing.T) {
	q := events.NewQueue()

	// Test append
	q.Append(big.NewInt(1))
	q.Append(big.NewInt(2))
	q.Append(big.NewInt(3))

	if q.Len() != 3 {
		t.Errorf("expected length of 3, got %d", q.Len())
	}

	if q.Head().Cmp(big.NewInt(1)) != 0 {
		t.Errorf("expected head of 1, got %d", q.Head())
	}

	if q.Tail().Cmp(big.NewInt(3)) != 0 {
		t.Errorf("expected tail of 3, got %d", q.Tail())
	}

	// Test pop
	if q.Pop().Cmp(big.NewInt(1)) != 0 {
		t.Error("expected 1 when popping")
	}

	if q.Pop().Cmp(big.NewInt(2)) != 0 {
		t.Error("expected 2 when popping")
	}

	if q.Pop().Cmp(big.NewInt(3)) != 0 {
		t.Error("expected 3 when popping")
	}

	if q.Pop() != nil {
		t.Error("expected nil when popping")
	}

}
