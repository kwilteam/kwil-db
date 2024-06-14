// Package main runs a node stress test tool with a few programs designed to
// impose a high load and test edge cases in transaction handling and dataset
// engine operations such as dataset deployment and action execution.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	host            string
	gatewayProvider bool
	key             string
	quiet           bool

	chainId string

	runTime time.Duration

	badgerInterval time.Duration
	viewInterval   time.Duration

	deployDropInterval time.Duration
	fastDropRate       int
	noDrop             bool

	noErrActs bool

	maxPosters   int
	postInterval time.Duration
	contentLen   int
	variableLen  bool

	txPollInterval time.Duration

	concurrentBroadcast bool
	nonceChaos          int
	rpcTiming           bool

	wg sync.WaitGroup
)

func main() {
	flag.StringVar(&host, "host", "http://127.0.0.1:8484", "provider's http url, schema is required")
	flag.BoolVar(&gatewayProvider, "gw", false, "gateway provider instead of vanilla provider, "+
		"need to make sure host is same as gateway's domain")
	flag.StringVar(&key, "key", "", "existing key to use instead of generating a new one")
	flag.BoolVar(&quiet, "q", false, "only print errors")

	flag.StringVar(&chainId, "chain", "", "chain ID to require (default is any)")

	flag.DurationVar(&runTime, "run", 30*time.Minute, "terminate after running this long")

	flag.DurationVar(&badgerInterval, "bi", -1, "badger kwild with read-only metadata requests at this interval")

	flag.DurationVar(&deployDropInterval, "ddi", -1, "deploy/drop datasets at this interval (but after drop tx confirms)")
	flag.IntVar(&fastDropRate, "ddn", 0, "immediately drop new dbs at a rate of 1/ddn (disable with <1)")
	flag.BoolVar(&noDrop, "nodrop", false, "don't drop the datasets deployed in the deploy/drop program")

	flag.BoolVar(&noErrActs, "ne", false, "don't make intentionally failed txns")

	flag.IntVar(&maxPosters, "ec", 4, "max concurrent unconfirmed action and procedure executions (to get multiple tx in a block), split between actions and procedures")
	flag.DurationVar(&postInterval, "ei", 10*time.Millisecond,
		"initiate non-view action execution at this interval (limited by max concurrency setting)")
	flag.DurationVar(&viewInterval, "vi", -1, "make view action call at this interval")
	flag.IntVar(&contentLen, "el", 50_000, "content length in an executed post action")
	flag.BoolVar(&variableLen, "vl", false, "pseudorandom variable content lengths, on (0,el]")

	flag.BoolVar(&concurrentBroadcast, "cb", false, "concurrent broadcast (do not wait for broadcast result before releasing nonce lock, will cause nonce errors due to goroutine race)")
	flag.IntVar(&nonceChaos, "nc", 0, "nonce chaos rate (apply nonce jitter every 1/nc times)")
	flag.BoolVar(&rpcTiming, "v", false, "print RPC durations")

	flag.DurationVar(&txPollInterval, "pollint", 400*time.Millisecond, "polling interval when waiting for tx confirmation")

	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	complete := errors.New("reached end time")
	ctx, cancel := context.WithTimeoutCause(context.Background(), runTime, complete)

	go func() {
		<-signalChan
		cancel()
	}()

	var exitCode int
	if err := hammer(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		exitCode = 1
	}

	cancel()
	wg.Wait()

	os.Exit(exitCode)
}
