package wallet

import (
	"fmt"
	"kwil/_archive/svcx/messaging/sub"
	"kwil/x"
	"kwil/x/async"
	"sync"
)

type confirmation_events struct {
	e       sub.TransientReceiver
	wg      sync.WaitGroup
	stop    chan x.Void
	done    chan x.Void
	mu      sync.Mutex
	handler func(ConfirmationEvent) async.Action
	status  int
}

func (c *confirmation_events) Start() error {
	c.mu.Lock()
	if c.status != 0 {
		c.mu.Unlock()
		return fmt.Errorf("already started")
	}

	c.status = 1
	c.mu.Unlock()

	err := c.e.Start()
	if err != nil {
		return err
	}

	go c.run()

	return nil
}

func (c *confirmation_events) Stop() error {
	c.mu.Lock()
	if c.status != 1 {
		c.mu.Unlock()
		return fmt.Errorf("confirmation event service is not running")
	}

	c.status = 2
	c.mu.Unlock()

	close(c.stop)
	c.e.Stop()

	return nil
}

func (c *confirmation_events) OnStop() <-chan x.Void {
	return c.done
}

func (c *confirmation_events) run() {
	done := false
	for !done {
		select {
		case <-c.stop:
			done = true
		case it := <-c.e.OnReceive():
			c.wg.Add(1)
			go c.handle_messages(it)
		}
	}

	c.wg.Wait()

	close(c.done)
}

func (c *confirmation_events) handle_messages(iter sub.MessageIterator) {
	if !iter.HasNext() {
		iter.Commit().WhenComplete(c.handle_if_error_and_set_wg_done)
		return
	}

	msg, _ := iter.Next()
	msg, request_id, err := decode_event(msg)
	if err != nil {
		fmt.Printf("error decoding event: %v", err)
		_ = c.Stop()
		return
	}

	ev := ConfirmationEvent{request_id, msg}
	c.handler(ev).
		OnCompleteA(&async.ContinuationA{
			Then:  c.get_next(iter),
			Catch: c.handle_if_error_and_set_wg_done,
		})
}

func (c *confirmation_events) get_next(iter sub.MessageIterator) func() {
	return func() {
		// iterate next now that event was handled
		c.handle_messages(iter)
	}
}

func (c *confirmation_events) handle_if_error_and_set_wg_done(err error) {
	if err == nil {
		c.wg.Done()
		return
	}

	fmt.Printf("error handling message: %v\n", err)
	c.wg.Done()
	_ = c.Stop()
}
