package rx

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

type TestStruct struct {
	value string
}

func Test_SetComplete(t *testing.T) {
	t.Run("Test_SetComplete", func(t *testing.T) {
		tc := &TestStruct{"done_test1"}
		task := NewTask[*TestStruct]()
		task.Complete(tc)
		v := task.Get()
		t.Log(v.value)
	})
}

func Test_SetError(t *testing.T) {
	task := NewTask[*TestStruct]()
	task.Fail(errors.New("test error"))
	rr := task.GetError()
	fmt.Println(rr)
	t.Log(task.GetError())
}

func Test_TaskOnSuccessOrError(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(4)

	tc := &TestStruct{"done_TaskOnSuccessOrError"}
	task := NewTask[*TestStruct]()

	task.Then(
		func(value *TestStruct) {
			t.Log(value.value)
			wg.Done()
		}).WhenComplete(
		func(value *TestStruct, err error) {
			t.Log(value.value)
			wg.Done()
		}).WhenComplete(
		func(value *TestStruct, err error) {
			t.Log(value.value)
			wg.Done()
		}).OnCompleteRunAsync(wg.Done)

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
			t.Log(value.value)
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
			t.Log(err)
			wg.Done()
		})

	task.Fail(errors.New("err_Test_TaskOnError"))

	wg.Wait()
}

func Test_SetCompleteNil(t *testing.T) {
	t.Run("Test_SetCompleteNil", func(t *testing.T) {
		task := NewTask[*TestStruct]()
		task.Complete(nil)
		v := task.Get()
		if v != nil {
			t.Fail()
		}
	})
}
