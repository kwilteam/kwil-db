package pub

import (
	"context"
	"fmt"
	. "kwil/x/cfgx"
	"kwil/x/messaging/mx"
	"kwil/x/utils"
	"os"
	"sync"
	"testing"
)

var test_msg = mx.RawMessage{
	Key:   []byte("test_key" + utils.GenerateRandomBase64String(10)),
	Value: []byte("test_value" + utils.GenerateRandomBase64String(10)),
}

func Test_Emitter_Basic(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	err := os.Setenv(Meta_Config_Test_Env, "../mx/test-meta-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	cfg := GetTestConfig().Select("messaging-emitter")
	topic := cfg.String("topic")
	if topic == "" {
		t.Fatal(fmt.Errorf("topic cannot be empty for test case"))
	}

	se := mx.SerdesByteArray()

	e, err := NewEmitter(cfg, se)
	if err != nil {
		t.Fatal(err)
	}

	msg := NewMessage(topic, test_msg, getAck[mx.RawMessage](t, wg))
	err = e.Send(context.Background(), msg)
	if err != nil {
		t.Error(err)
	}

	wg.Wait()
	e.Close()
}

func getAck[T any](t *testing.T, wg *sync.WaitGroup) AckNackFn {
	return AckNack(func(e error) {
		if e != nil {
			t.Fatal(e)
		} else {
			t.Log("Message sent")
		}
		wg.Done()
	})
}
