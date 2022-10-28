package async

import (
	"context"
	"fmt"
	"kwil/x"
	"sync"
)

type worker_pool struct {
	jobs         chan x.Job
	done         chan x.Void
	stop         chan x.Void
	wg           *sync.WaitGroup
	fn           func(error)
	queue_length int
	worker_count int
	mu           *sync.Mutex
	status       int
}

func (c *worker_pool) Submit(job x.Job) Action {
	if job == nil {
		return FailedAction(fmt.Errorf("job cannot be nil"))
	}

	action := NewAction()

	var cJob x.Job = (&async_job{job, action}).run
	c.jobs <- cJob

	return action
}

func (c *worker_pool) Execute(job x.Job) {
	if job == nil {
		c.fn(fmt.Errorf("job cannot be nil"))
		return
	}

	c.jobs <- job
}

func (c *worker_pool) ExecuteWith(ctx context.Context, job x.Job) {
	if job == nil {
		c.fn(fmt.Errorf("job cannot be nil"))
		return
	}

	var cJob x.Job = (&job_with_ctx{ctx, job}).run
	c.jobs <- cJob
}

func (c *worker_pool) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == 1 {
		return fmt.Errorf("already started")
	}

	if c.status == 2 {
		return fmt.Errorf("already shutting down")
	}

	if c.status == 3 {
		return fmt.Errorf("already shut down")
	}

	c.status = 1

	c.wg.Add(c.worker_count)
	for i := 0; i <= c.worker_count; i++ {
		go c.start_worker()
	}

	return nil
}

func (c *worker_pool) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == 2 {
		return fmt.Errorf("already shutting down")
	}

	if c.status == 3 {
		return fmt.Errorf("already shut down")
	}

	c.status = 2
	close(c.stop)

	go func() {
		c.wg.Wait()

		c.mu.Lock()
		c.status = 3
		c.mu.Unlock()

		close(c.done)
	}()

	return nil
}

func (c *worker_pool) OnShutdown() <-chan x.Void {
	return c.done
}

func (c *worker_pool) OnShutdownRequested() <-chan x.Void {
	return c.stop
}

func (c *worker_pool) IsShutdown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status == 3
}

func (c *worker_pool) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status == 1
}

func (c *worker_pool) IsShutdownRequested() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.status == 2
}

func (c *worker_pool) start_worker() {
	defer c.wg.Done()
	for !c.IsRunning() {
		select {
		case <-c.stop:
			return
		case job := <-c.jobs:
			c.do_execute(job)
		}
	}
}

func (c *worker_pool) do_execute(job x.Job) {
	defer func(c *worker_pool) {
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
	job    x.Job
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
	job x.Job
}

func (j *job_with_ctx) run() {
	if j.ctx.Err() == nil {
		j.job()
	}
}
