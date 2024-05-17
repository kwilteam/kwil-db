package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kwilteam/kwil-db/core/log"
)

var (
	statsFile string

	rpcServers string
)

func main() {
	// Flag support for stats file name, rpc servers to query.
	flag.StringVar(&statsFile, "output", "stats.json", "stats file name")
	flag.StringVar(&rpcServers, "rpcservers", "http://localhost:26657", "comma separated list of rpc servers to query stats from")

	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	logger := log.New(log.Config{
		Level:       log.InfoLevel.String(),
		OutputPaths: []string{"stdout"},
		Format:      log.FormatPlain,
		EncodeTime:  log.TimeEncodingEpochMilli, // for readability, log.TimeEncodingRFC3339Milli
	})

	addresses := strings.Split(rpcServers, ",")
	statsMonitor, err := newStatsMonitor(addresses, statsFile, logger)
	if err != nil {
		logger.Error("failed to create stats monitor", log.Error(err))
		os.Exit(1)
	}

	err = statsMonitor.Run(signalChan)
	if err != nil {
		logger.Error("failed to run stats monitor", log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}
