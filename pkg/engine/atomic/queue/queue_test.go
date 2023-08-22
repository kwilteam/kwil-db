package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/engine/atomic/queue"
)

func Test_Queue(t *testing.T) {
	ctx := context.Background()

	q := queue.New(10)

	done, err := q.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		done()
	}()

	done2, err := q.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	done2()
}

func Test_Queue_Full(t *testing.T) {
	ctx := context.Background()

	q := queue.New(1)

	done, err := q.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}

	errChan := make(chan error)
	// one of these should wait for the above to finish, the other should return ErrQueueFull
	go func() {
		done2, err := q.Wait(ctx)
		if err != nil {
			errChan <- err
			return
		}
		done2()
	}()
	go func() {
		done2, err := q.Wait(ctx)
		if err != nil {
			errChan <- err
			return
		}
		done2()
	}()

	err = <-errChan
	if err != queue.ErrQueueFull {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	done()
}
