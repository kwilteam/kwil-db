package messaging

import (
	"context"
	"kwil/x/messaging/config"
	"kwil/x/utils"
	"sync"
	"testing"
)

var test_msg = RawMessage{
	Key:   []byte("test_key" + utils.GenerateRandomBase64String(10)),
	Value: []byte("test_value" + utils.GenerateRandomBase64String(10)),
}

func Test_Producer_Basic(t *testing.T) {
	cfg := config.GetTestConfig()
	pCfg := cfg.Select("messaging-producer-service-test")
	p, err := NewProducer(pCfg, SerdesByteArray)
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	ctx := context.Background()
	msg := MessageP(test_msg, getAck(t, p, wg))

	err = p.Submit(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
}

func getAck[T any](t *testing.T, p Producer[T], wg *sync.WaitGroup) AckNackFn {
	return AckNack(func(e error) {
		if e != nil {
			t.Fatal(e)
		} else {
			t.Log("Message sent")
		}

		p.Close()
		wg.Done()
	})
}
