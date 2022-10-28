package async

import (
	"fmt"
	"kwil/x"
	"sync"
)

func NewSingleWorker(opts ...*OptPool) (WorkerPool, error) {
	return NewWorkerPool(1, opts...)
}

func NewWorkerPool(workerCount int, opts ...*OptPool) (WorkerPool, error) {
	if workerCount <= 0 {
		return nil, fmt.Errorf("workerCount must be greater than 0")
	}

	e := &worker_pool{}
	for _, opt := range opts {
		opt.configure(e)
	}

	if e.queue_length < workerCount {
		return nil, fmt.Errorf("queue_length must be greater than or equal to workerCount")
	}

	if e.fn == nil {
		e.fn = func(err error) {
			// TODO: use logger here
			fmt.Println(err)
		}
	}

	e.jobs = make(chan x.Job, e.queue_length)
	e.done = make(chan x.Void)
	e.stop = make(chan x.Void)
	e.wg = &sync.WaitGroup{}
	e.mu = &sync.Mutex{}
	e.worker_count = workerCount

	return e, nil
}

var immediateExecutor x.Executor = &immediate_executor{}
var asyncExecutor x.Executor = &async_executor{}

// DefaultExecutor returns an executor that will execute the given
// function asynchronously using the native go scheduler (e.g., go func(){}).
func DefaultExecutor() x.Executor {
	return asyncExecutor
}

// ImmediateExecutor will execute the given function immediately on the
// callers thread.
func ImmediateExecutor() x.Executor {
	return immediateExecutor
}

type async_executor struct{}

func (e *async_executor) Execute(job x.Job) {
	go job()
}

type immediate_executor struct{}

func (e *immediate_executor) Execute(job x.Job) {
	job()
}
