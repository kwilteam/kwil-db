package retry_test

import (
	"context"
	"fmt"
	"kwil/pkg/log"
	"kwil/pkg/utils/retry"
	"testing"
	"time"
)

// WARNING: the unit tests in here take about 12 seconds to run

func Test_Retry(t *testing.T) {
	strct := &TestStruct{
		counter: 0,
	}
	retrier := retry.New(strct)

	ctx := context.Background()
	err := retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		return t.TestMethod()
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		_, err := t.TestMethod2()
		return err
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// I am mostly just eyeballing the time this test takes
func Test_Retry_Opts(t *testing.T) {
	strct := &TestStruct{
		counter: 0,
	}
	retrier := retry.New(strct,
		retry.WithFactor[*TestStruct](2),
		retry.WithMax[*TestStruct](time.Millisecond*5000),
		retry.WithMin[*TestStruct](time.Millisecond*100),
		retry.WithoutJitter[*TestStruct](),
		retry.WithLogger[*TestStruct](log.New(log.Config{
			Level:       "debug",
			OutputPaths: []string{"stdout"},
		})),
	)

	ctx := context.Background()
	err := retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		return t.TestMethod()
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	err = retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		_, err := t.TestMethod2()
		return err
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func Test_Retry_Timeout(t *testing.T) {
	strct := &TestStruct{
		counter: 0,
	}
	retrier := retry.New(strct)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	go func(cancel context.CancelFunc) {
		time.Sleep(time.Second * 1)
		cancel()
	}(cancel)

	err := retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		_, err := t.TestMethod2()
		return err
	})
	if err != context.Canceled {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func Test_Exceeding_Max(t *testing.T) {
	// test with 1 less than counter as the max retry, should fail
	strct := &TestStruct{
		counter: 0,
	}
	retrier := retry.New(strct,
		retry.WithMax[*TestStruct](time.Millisecond*1000),
		retry.WithMaxRetries[*TestStruct](2),
	)

	ctx := context.Background()
	err := retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		_, err := t.TestMethod2()
		return err
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	// test with 1 more than counter as the max retry, should succeed
	strct = &TestStruct{
		counter: 0,
	}
	retrier = retry.New(strct,
		retry.WithMax[*TestStruct](time.Millisecond*1000),
		retry.WithMaxRetries[*TestStruct](3),
	)

	err = retrier.Retry(ctx, func(_ context.Context, t *TestStruct) error {
		_, err := t.TestMethod2()
		return err
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

type TestStruct struct {
	counter int
}

func (t *TestStruct) TestMethod() error {
	return nil
}

func (t *TestStruct) TestMethod2() (string, error) {
	if t.counter == 3 {
		return "success", nil
	}

	t.counter++
	return "", fmt.Errorf("error")
}
