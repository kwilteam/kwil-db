package syncx

import (
	"fmt"
	"kwil/x"
	"kwil/x/async"
	"sync"
	"sync/atomic"
	"testing"
)

func Test_Chan_Close_Basics(t *testing.T) {
	ch := NewChanBuffered[x.Void](10)

	m := sync.Mutex{}
	m.Lock()
	if !ch.Write(x.Void{}) {
		t.Fail()
	}

	if !ch.Write(x.Void{}) {
		t.Fail()
	}

	ch.Close()

	if ch.Write(x.Void{}) {
		t.Fail()
	}
}

func Test_Chan_Drain(t *testing.T) {
	ch := NewChanBuffered[x.Void](10)
	c := async.NewAction()

	go func() {
		if !ch.Write(x.Void{}) {
			t.Fail()
		}

		if !ch.Write(x.Void{}) {
			t.Fail()
		}

		ch.Close()

		if ch.Write(x.Void{}) {
			t.Fail()
		}

		c.Complete()
	}()

	<-c.Then(func() {
		el, err := ch.Drain(nil)
		if err != nil {
			t.Fail()
		}

		if len(el) != 2 {
			t.Fail()
		}
	}).DoneCh()
}

func Test_Chan_Read(t *testing.T) {
	ch := NewChanBuffered[x.Void](50)

	go func() {
		for i := 0; i < 100; i++ {
			if !ch.Write(x.Void{}) {
				t.Fail()
			}
		}

		ch.Close()
	}()

	cnt := 0
	done := false
	for !done {
		select {
		case _, ok := <-ch.Read():
			if !ok {
				done = true
			} else {
				cnt++
			}
		case <-ch.ClosedCh():
			done = true
		}
	}

	el, err := ch.Drain(nil)
	if err != nil {
		t.Fail()
	}

	if len(el) != 100-cnt {
		t.Fail()
	}
}

func Test_Chan_Many_Writers(t *testing.T) {
	max := 20000
	ch := NewChanBuffered[x.Void](max)

	ready := &sync.WaitGroup{}
	ready.Add(8)

	readReady := &sync.WaitGroup{}
	readReady.Add(1)

	readDone := &sync.WaitGroup{}
	readDone.Add(1)

	write_cnt := int32(0)
	for i := 0; i < 8; i++ {
		go func() {
			ready.Done()
			readReady.Wait()
			for i := 0; i < max; i++ {
				if !ch.Write(x.Void{}) {
					break
				}
				atomic.AddInt32(&write_cnt, 1)
			}
		}()
	}

	read_cnt := int32(0)
	go func() {
		ready.Wait()
		readReady.Done()
		halfway := int32(8 * max / 3)
		done := false
		for !done {
			select {
			case _, ok := <-ch.Read():
				if !ok {
					done = true
				} else {
					if halfway == atomic.AddInt32(&read_cnt, 1) {
						ch.Close()
					}
				}
			case <-ch.LockCh():
				done = true
			}
		}
		readDone.Done()
	}()

	readDone.Wait()

	el, err := ch.Drain(nil)
	if err != nil {
		t.Fail()
	}

	r_cnt := int(atomic.LoadInt32(&read_cnt))
	w_cnt := int(atomic.LoadInt32(&write_cnt))

	if len(el) != w_cnt-r_cnt {
		t.Fail()
	}

	fmt.Printf("read_cnt: %d, write_cnt: %d, drain_cnt: %d\n", r_cnt, w_cnt, len(el))
}
