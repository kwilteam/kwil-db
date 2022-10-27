package async

import (
	"context"
	"fmt"
	"kwil/x"
	"sync"
)

type ExecutorService interface {
	Execute(job Job)
	ExecuteWith(ctx context.Context, job Job)
	Submit(job Job) Action
	Stop()
	OnStop() <-chan x.Void
	IsStopped() bool
}

type executor_service struct {
	jobs   chan Job
	done   chan x.Void
	stop   chan x.Void
	wg     *sync.WaitGroup
	fn     func(error)
	mu     *sync.Mutex
	status int
}

func NewExecutorService(workerCount int, unhandledErrorHandler func(error)) (ExecutorService, error) {
	if workerCount <= 0 {
		panic("workerCount must be greater than 0")
	}

	jobs := make(chan Job, 10*workerCount)
	stop := make(chan x.Void)
	wg := &sync.WaitGroup{}
	wg.Add(workerCount)

	if unhandledErrorHandler == nil {
		unhandledErrorHandler = func(err error) {
			// TODO: use logger here
			fmt.Println(err)
		}
	}

	e := &executor_service{
		jobs,
		make(chan x.Void),
		stop,
		wg,
		unhandledErrorHandler,
		&sync.Mutex{},
		0}

	for i := 0; i <= workerCount; i++ {
		go e.start_worker()
	}

	return e, nil
}

func (c *executor_service) Submit(job Job) Action {
	if job == nil {
		return FailedAction(fmt.Errorf("job cannot be nil"))
	}

	action := NewAction()

	var cJob Job = (&async_job{job, action}).run
	c.jobs <- cJob

	return action
}

func (c *executor_service) Execute(job Job) {
	if job == nil {
		c.fn(fmt.Errorf("job cannot be nil"))
		return
	}

	c.jobs <- job
}

func (c *executor_service) ExecuteWith(ctx context.Context, job Job) {
	if job == nil {
		c.fn(fmt.Errorf("job cannot be nil"))
		return
	}

	var cJob Job = (&job_with_ctx{ctx, job}).run
	c.jobs <- cJob
}

func (c *executor_service) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status > 0 {
		return
	}

	c.status = 1
	close(c.stop)

	go func() {
		c.wg.Wait()

		c.mu.Lock()
		c.status = 2
		c.mu.Unlock()

		close(c.done)
	}()
}

func (c *executor_service) OnStop() <-chan x.Void {
	return c.done
}

func (c *executor_service) IsStopped() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status == 2
}

func (c *executor_service) start_worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.stop:
			return
		case job := <-c.jobs:
			c.do_execute(job)
		}
	}
}

func (c *executor_service) do_execute(job Job) {
	defer func(c *executor_service) {
		if r := recover(); r != nil {
			e, ok := r.(error)
			var err error
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
			c.fn(err)
		}
	}(c)

	job()
}

type async_job struct {
	job    Job
	action Action
}

func (j *async_job) run() {
	if j.action.IsDone() {
		return // job was set by caller of Submit
	}

	defer func(j *async_job) {
		if r := recover(); r != nil {
			e, ok := r.(error)
			var err error
			if !ok {
				err = fmt.Errorf("unknown panic: %v", r)
			} else {
				err = fmt.Errorf("panic: %v", e)
			}
			j.action.Fail(err)
		}
	}(j)

	j.job()

	j.action.Complete()
}

type job_with_ctx struct {
	ctx context.Context
	job Job
}

func (j *job_with_ctx) run() {
	if j.ctx.Err() == nil {
		j.job()
	}
}
