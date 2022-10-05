package syncx

import (
	"kwil/x/rx"
	"testing"
)

func Test_Chan_Close_Basics(t *testing.T) {
	ch := NewChanBuffered[rx.Void](10)

	if !ch.Write(rx.Void{}) {
		t.Fail()
	}

	if !ch.Write(rx.Void{}) {
		t.Fail()
	}

	ch.Close()

	if ch.Write(rx.Void{}) {
		t.Fail()
	}
}

func Test_Chan_Drain(t *testing.T) {
	ch := NewChanBuffered[rx.Void](10)
	c := rx.NewContinuation()

	go func() {
		if !ch.Write(rx.Void{}) {
			t.Fail()
		}

		if !ch.Write(rx.Void{}) {
			t.Fail()
		}

		ch.Close()

		if ch.Write(rx.Void{}) {
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
	}).DoneChan()
}

func Test_Chan_Read(t *testing.T) {
	ch := NewChanBuffered[rx.Void](50)

	go func() {
		for i := 0; i < 100; i++ {
			if !ch.Write(rx.Void{}) {
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
