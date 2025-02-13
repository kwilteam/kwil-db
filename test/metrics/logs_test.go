package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_LogExport(t *testing.T) {
	blockStore := "../../.net/node0/blockstore"
	startBlock := 10927
	endBlock := 11000

	logFile := "logs.csv"
	logs, err := ExtractLogs(blockStore, int64(startBlock), int64(endBlock), logFile)
	require.NoError(t, err)

	// Check that the log file was created
	_, err = os.Stat(logFile)
	require.NoError(t, err)

	// analyze the logs
	metricsFile := "metrics.csv"
	metrics := AnalyzeLogs(logs, TestParams{
		ProposeTimeout: 1 * time.Second,
		MaxBlockSize:   6 * 1024 * 1024,
		Concurrency:    500,
		PayloadSize:    10000,
	})
	WriteMetrics(metrics, metricsFile, "csv")
}
