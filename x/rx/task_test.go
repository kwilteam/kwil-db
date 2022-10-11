package rx

import (
	"errors"
	"sync"
	"testing"
)

type TestStruct struct {
	value string
}

func Test_SetComplete(t *testing.T) {
	tc := &TestStruct{"done_test1"}
	task := NewTask[*TestStruct]()
	task.Complete(tc)
	v := task.Get()
	if v.value != tc.value {
		t.Fail()
	}
}

func Test_SetError(t *testing.T) {
	task := NewTask[*TestStruct]()
	err := errors.New("test error")
	task.Fail(err)
	if err != task.GetError() {
		t.Fail()
	}
}

func Test_Order(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(4)

	tc := &TestStruct{"done_TaskOnSuccessOrError"}
	task := NewTask[*TestStruct]()

	count := 4

	task.Then(
		func(value *TestStruct) {
			if count != 1 {
				t.Fail()
			}
			count--
			wg.Done()
		}).WhenComplete(
		func(value *TestStruct, err error) {
			if count != 2 {
				t.Fail()
			}
			count--
			wg.Done()
		}).WhenComplete(
		func(value *TestStruct, err error) {
			if count != 3 {
				t.Fail()
			}
			count--
			wg.Done()
		}).OnComplete(&ContinuationT[*TestStruct]{
		Finally: func() {
			if count != 4 {
				t.Fail()
			}
			count--
			wg.Done()
		}})

	task.Complete(tc)

	wg.Wait()
}

func Test_WhenComplete(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tc := &TestStruct{"done_TaskOnSuccess"}
	task := NewTask[*TestStruct]()

	task.ThenCatchFinally(&ContinuationT[*TestStruct]{
		Then: func(value *TestStruct) {
			if value.value != tc.value {
				t.Fail()
			}
			wg.Done()
		},
	})

	task.Complete(tc)

	wg.Wait()
}

func Test_Catch(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	task.Catch(
		func(err error) {
			wg.Done()
		})

	task.Fail(errors.New("Test_Catch"))

	wg.Wait()
}

func Test_Complete(t *testing.T) {
	task := NewTask[string]()
	task2 := task
	task.Complete("Test_Complete")
	if task.Get() != task2.Get() {
		t.Fail()
	}
}

func Test_ThenCatchFinally(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(4)

	task := NewTask[*TestStruct]()

	task.ThenCatchFinally(&ContinuationT[*TestStruct]{
		Then: func(v *TestStruct) {
			wg.Done()
		},
		Catch: func(err error) {
			t.Fail()
		},
		Finally: func() {
			wg.Done()
		},
	})

	go func() {
		task.Complete(&TestStruct{"test"})
	}()

	task2 := NewTask[*TestStruct]()

	count := 2
	task2.ThenCatchFinally(&ContinuationT[*TestStruct]{
		Then: func(v *TestStruct) {
			t.Fail()
		},
		Catch: func(err error) {
			if count != 2 {
				t.Fail()
			}
			wg.Done()
			count--
		},
		Finally: func() {
			if count != 1 {
				t.Fail()
			}
			wg.Done()
		},
	})

	go func() {
		task2.Fail(errors.New("force fail"))
	}()

	wg.Wait()
}

func Test_AsAsync(t *testing.T) {
	tc := &TestStruct{"Test_AsAsync"}
	task := NewTask[*TestStruct]()

	d := []int{0}
	task.AsAsync(nil).Then(
		func(value *TestStruct) {
			d[0] = 1
		})

	task.Complete(tc)
	if d[0] == 1 {
		t.Fail()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	task2 := NewTask[*TestStruct]()

	d2 := []int{0}
	task2.AsAsync(nil).Then(
		func(value *TestStruct) {
			d2[0] = 1
			wg.Done()
		})

	task2.Complete(tc)
	if d2[0] == 1 {
		t.Fail()
	}

	wg.Wait()

	if d2[0] != 1 {
		t.Fail()
	}

	task3 := NewTask[*TestStruct]()

	d3 := []int{0}
	task3.AsAsync(ImmediateExecutor()).Then(
		func(value *TestStruct) {
			d3[0] = 1
		})

	task3.Complete(tc)
	if d3[0] != 1 {
		t.Fail()
	}
}

func Test_Then(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	d := []int{0}
	task.Then(func(value *TestStruct) {
		if value != nil {
			t.Fail()
		}
		d[0] = 1
		wg.Done()
	})

	task.Complete(nil)

	if d[0] != 1 {
		t.Fail()
	}

	wg.Wait()
}

func Test_Compose(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	d := []int{0}
	outer := task.Compose(func(value *TestStruct, err error) Task[*TestStruct] {
		if value != nil {
			t.Fail()
		}

		if err != nil {
			t.Error(err)
			return Failure[*TestStruct](err)
		}

		return Call(func() (*TestStruct, error) {
			d[0] = 1
			defer wg.Done()
			return &TestStruct{"inner"}, nil
		})
	})

	task.Complete(nil)

	if d[0] == 1 {
		t.Fail()
	}

	wg.Wait()

	if d[0] != 1 {
		t.Fail()
	}

	if outer.Get() == nil || outer.Get().value != "inner" {
		t.Fail()
	}
}

func Test_ComposeToErr(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	d := []int{0}
	outer := task.Compose(func(value *TestStruct, err error) Task[*TestStruct] {
		if value != nil {
			t.Fail()
		}

		if err != nil {
			t.Fail()
		}

		return Call[*TestStruct](func() (*TestStruct, error) {
			d[0] = 1
			defer wg.Done()
			return nil, errors.New("inner")
		})
	})

	task.Complete(nil)

	if d[0] == 1 {
		t.Fail()
	}

	wg.Wait()

	if d[0] != 1 {
		t.Fail()
	}

	if outer.GetError() == nil || outer.GetError().Error() != "inner" {
		t.Fail()
	}
}

func Test_Handle(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	d := []int{0}
	outer := task.Handle(func(value *TestStruct, err error) (*TestStruct, error) {
		if value != nil {
			t.Fail()
		}

		if err != nil {
			t.Fail()
		}

		d[0] = 1
		defer wg.Done()

		return &TestStruct{"inner"}, nil
	})

	task.Complete(nil)

	if d[0] != 1 {
		t.Fail()
	}

	wg.Wait()

	if outer.Get() == nil || outer.Get().value != "inner" {
		t.Fail()
	}
}

func Test_HandleErr(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	d := []int{0}
	outer := task.Handle(func(value *TestStruct, err error) (*TestStruct, error) {
		if value != nil {
			t.Fail()
		}

		if err != nil {
			t.Fail()
		}

		d[0] = 1
		defer wg.Done()

		return nil, errors.New("inner")
	})

	task.Complete(nil)

	if d[0] != 1 {
		t.Fail()
	}

	wg.Wait()

	if outer.GetError() == nil || outer.GetError().Error() != "inner" {
		t.Fail()
	}
}
