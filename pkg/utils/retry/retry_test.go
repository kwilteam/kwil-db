package retry_test

import (
	"context"
	"fmt"
	"kwil/pkg/log"
	retry "kwil/pkg/utils/retry"
	"testing"
	"time"
)

func Test_Retry(t *testing.T) {
	strct := &Struct1{
		counter: 0,
	}

	// expect this to run until it passes
	err := retry.Retry(func() error {
		return strct.TestMethod()
	},
		retry.WithFactor(3),
		retry.WithMax(time.Millisecond*500),
		retry.WithMin(time.Millisecond*100),
		retry.WithLogger(log.New(log.Config{
			Level:       "debug",
			OutputPaths: []string{"stdout"},
		})),
		retry.WithoutJitter(),
	)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	strct.Reset()
	//set max retries to 3, should work
	err = retry.Retry(func() error {
		return strct.TestMethod()
	},
		retry.WithFactor(3),
		retry.WithMax(time.Millisecond*500),
		retry.WithMin(time.Millisecond*100),
		retry.WithMaxRetries(3),
	)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	strct.Reset()
	//set max retries to 2, should fail
	err = retry.Retry(func() error {
		return strct.TestMethod()
	},
		retry.WithFactor(3),
		retry.WithMax(time.Millisecond*500),
		retry.WithMin(time.Millisecond*100),
		retry.WithMaxRetries(2),
	)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// try context timeout
	strct.Reset()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	go func(cancel context.CancelFunc) {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}(cancel)

	err = retry.Retry(func() error {
		return strct.TestMethod()
	},
		retry.WithFactor(3),
		retry.WithMax(time.Millisecond*500),
		retry.WithMin(time.Millisecond*100),
		retry.WithContext(ctx),
	)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

}

type Struct1 struct {
	counter int
}

func (s *Struct1) TestMethod() error {
	if s.counter == 3 {
		return nil
	}
	s.counter++
	return fmt.Errorf("error")
}

func (s *Struct1) Reset() {
	s.counter = 0
}
