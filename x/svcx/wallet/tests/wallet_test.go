package tests

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"kwil/x"
	"kwil/x/async"
	"kwil/x/cfgx"
	"kwil/x/svcx/messaging/mx"
	"kwil/x/svcx/wallet"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func Test_Wallet_Single_Request(t *testing.T) {
	// uncomment to test (not meant for build based unit testing)
	//run_test(1, t)
}

func Test_Wallet_Multi_Request(t *testing.T) {
	// uncomment to test (not meant for build based unit testing)
	//run_test(1000, t)
}

func run_test(cnt int, t *testing.T) {
	//err := os.Setenv(cfgx.Root_Dir_Env, "./tests")
	//if err != nil {
	//	t.Fatal(err)
	//}

	// Below confirmed *working* on first message for wallet service
	// TODO: look at issue in processing service
	walletService, err := loadWalletService()
	if err != nil {
		fmt.Println("failed to load wallet service: ", err)
		t.Error(err)
		return
	}

	time.Sleep(10 * time.Millisecond) // making sure things are started before getting the start time

	started := time.Now()

	wg := &sync.WaitGroup{}
	wg.Add(cnt)
	var requests []async.Listenable[x.Void]
	for i := 0; i < cnt; i++ {
		requests = append(requests,
			walletService.
				Submit(newMessage(i)).
				WhenComplete(onSubmit(wg, i)))
	}

	select {
	case <-time.After(30 * time.Minute):
		t.Error(fmt.Errorf("timeout awaiting responses"))
		return
	case <-async.All(requests...).DoneCh():
		fmt.Printf("completed in %v\n", time.Since(started))
	}

	wg.Wait()

	millis := time.Since(started) / time.Millisecond
	fmt.Printf("total %d milliseconds for %d messages with an average of %d milliseconds per message\n", millis, cnt, int64(millis)/int64(cnt))
}

func loadWalletService() (wallet.RequestService, error) {
	p, err := wallet.NewRequestProcessor(cfgx.GetConfig(), wallet.SyncTransform(func(msg *mx.RawMessage) (*mx.RawMessage, error) {
		key := string(msg.Key[:])
		if !strings.HasPrefix(key, "key__") {
			fmt.Println("unknown message type during test")
		}

		return msg, nil
	}))

	if err != nil {
		return nil, err
	}

	w, err := wallet.NewRequestService(cfgx.GetConfig())
	if err != nil {
		return nil, err
	}

	err = p.Start()
	if err != nil {
		return nil, err
	}

	err = w.Start()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func newMessage(id int) (context.Context, *mx.RawMessage) {
	return nil, &mx.RawMessage{
		Key:   []byte("key__" + strconv.Itoa(id)),
		Value: []byte("payload__" + uuid.New().String()),
	}
}

func onSubmit(wg *sync.WaitGroup, id int) func(error) {
	return func(err error) {
		if err != nil {
			fmt.Printf("err [%d]: %v\n", id, err)
		}
		wg.Done()
	}
}
