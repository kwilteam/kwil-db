package messaging

import (
	"context"
	"kwil/x/messaging/config"
	"kwil/x/rx"
	"kwil/x/utils"
	"sync"
	"testing"
)

func Test_Producer_Basic(t *testing.T) {
	cfg := config.GetTestConfig()
	pCfg := cfg.Select("messaging-producer-service-test")
	p, err := NewProducer(pCfg, SerdesByteArray)
	if err != nil {
		t.Fatal(err)
	}

	msg := &RawMessage{
		Key:   []byte("test_key" + utils.GenerateRandomBase64String(10)),
		Value: []byte("test_value" + utils.GenerateRandomBase64String(10)),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	p.Submit(context.Background(), msg).WhenCompleteInvoke(&rx.CompletionC{
		Then: func() {
			t.Log("Message sent")
		},
		Catch: func(err error) {
			t.Fatal(err)
		},
		Finally: func() {
			p.Close()
			wg.Done()
		},
	})

	wg.Wait()
}
