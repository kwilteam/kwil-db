package pub

import (
	"context"
	"fmt"
	. "kwil/x/cfgx"
	"kwil/x/messaging/mx"
	"kwil/x/utils"
	"os"
	"testing"
)

var test_msg = mx.RawMessage{
	Key:   []byte("test_key" + utils.GenerateRandomBase64String(10)),
	Value: []byte("test_value" + utils.GenerateRandomBase64String(10)),
}

func Test_Emitter_Sync(t *testing.T) {
	if t == nil {
		return // intentionally ignore this test for normal ops
	}

	err := os.Setenv(Meta_Config_Test_Env, "../mx/test-meta-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	cfg := GetTestConfig().Select("messaging-emitter")
	topic := cfg.String("default-topic")
	if topic == "" {
		t.Fatal(fmt.Errorf("default-topic cannot be empty for test case"))
	}

	e, err := NewEmitterSingleClient(cfg, mx.SerdesByteArray())
	if err != nil {
		t.Fatal(err)
	}

	// ToDO: Need actual message
	err = e.SendSync(context.Background(), test_msg)
	if err != nil {
		t.Error(err)
	}

	if e != nil {
		t.Fatal(e)
	} else {
		t.Log("Message sent")
	}

	e.Close()
}

func Test_Emitter_Async(t *testing.T) {
	if t == nil {
		return // intentionally ignore this test for normal ops
	}

	err := os.Setenv(Meta_Config_Test_Env, "../mx/test-meta-config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	cfg := GetTestConfig().Select("messaging-emitter")
	topic := cfg.String("default-topic")
	if topic == "" {
		t.Fatal(fmt.Errorf("default-topic cannot be empty for test case"))
	}

	e, err := NewEmitterSingleClient(cfg, mx.SerdesByteArray())
	if err != nil {
		t.Fatal(err)
	}

	// ToDO: Need actual message
	ctx := context.Background()
	<-e.Send(ctx, test_msg).
		WhenComplete(func(err error) {
			if err != nil {
				t.Error(err)
			} else {
				t.Log("Message sent")
			}
		}).DoneCh()

	e.Close()
}
