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

func Test_TaskOnSuccessOrError(t *testing.T) {
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
		}).OnComplete(&Completion[*TestStruct]{
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

func Test_TaskOnSuccess(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	tc := &TestStruct{"done_TaskOnSuccess"}
	task := NewTask[*TestStruct]()

	task.AsAsync().Then(
		func(value *TestStruct) {
			wg.Done()
		})

	task.Complete(tc)

	wg.Wait()
}

func Test_TaskOnError(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	task := NewTask[*TestStruct]()

	task.Catch(
		func(err error) {
			wg.Done()
		})

	task.Fail(errors.New("err_Test_TaskOnError"))

	wg.Wait()
}

func Test_TaskCopyResultAreSame(t *testing.T) {
	task := NewTask[string]()
	task2 := task
	task.Complete("test_copy")
	if task.Get() != task2.Get() {
		t.Fail()
	}
}
