package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// This package includes requires tooling for extracting logs from the blockstore
// given a start and end block height and writes the logs to a csv file.
// The logs can then be analyzed to extract metrics such as Transactions Per Second (TPS),
// Block Rate, Data Ingress Rate, Throughput Utilization, Average Time to Finality etc.

var (
	logFileName     = "logs.csv"
	metricsFileName = "metrics"
	resultsDirName  = "results"
	blockstore      = "blockstore"
	startBlockFile  = "start_block.txt"
	endBlockFile    = "end_block.txt"
	testParamsFile  = "test_params.json"
)

func main() {
	var resultsDir, fileFormat, logDir string

	flag.StringVar(&resultsDir, "output", resultsDirName, "Directory to write the analyzed metrics and logs to")
	flag.StringVar(&fileFormat, "format", "json", "File format to write the analyzed metrics to. Supported formats: csv, json")
	flag.StringVar(&logDir, "logs", "", "Directory containing the performance test logs to analyze. Should include blockstore, start and end block files, test params ")

	flag.Parse()

	// Ensure that the logDir contains all the required files to analyze the test run
	// - blockstore
	// - start and end block files
	// - test params file

	if err := verifyLogDir(logDir); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// get the start, end block heights and test params
	startBlock, err := getBlockHeight(filepath.Join(logDir, startBlockFile))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	endBlock, err := getBlockHeight(filepath.Join(logDir, endBlockFile))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	testParams, err := getTestParams(filepath.Join(logDir, testParamsFile))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// create the results dir
	if err := os.MkdirAll(resultsDir, os.ModePerm); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Extract the logs from the blockstore
	bs := filepath.Join(logDir, blockstore)
	logFile := filepath.Join(resultsDir, logFileName)
	logs, err := ExtractLogs(bs, startBlock, endBlock, logFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Write the logs to a file
	if err := WriteLogs(logs, logFile); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Analyze the logs and extract metrics
	metrics := AnalyzeLogs(logs, *testParams)
	metricsFile := filepath.Join(resultsDir, metricsFileName+"."+fileFormat)
	// Write the metrics to a file
	if err := WriteMetrics(metrics, metricsFile, fileFormat); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Metrics written to %s\n", metricsFile)
	fmt.Printf("Logs written to %s\n", logFile)
}

func getBlockHeight(filename string) (int64, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %v", err)
	}
	heightStr := strings.Trim(string(data), "\n")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %v", err)
	}
	return height, nil
}

func getTestParams(filename string) (*TestParams, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var params TestParams
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test params: %v", err)
	}
	return &params, nil
}

func verifyLogDir(logDir string) error {
	// ensure that the dir exists
	if exists, err := exists(logDir); !exists || err != nil {
		return fmt.Errorf("log dir does not exist: %v, %v", logDir, err)
	}

	// blockstore
	bs := filepath.Join(logDir, blockstore)
	if exists, err := exists(bs); !exists || err != nil {
		return fmt.Errorf("blockstore does not exist: %v, %v", bs, err)
	}

	// start and end block files
	sb := filepath.Join(logDir, startBlockFile)
	eb := filepath.Join(logDir, endBlockFile)
	if exists, err := exists(sb); !exists || err != nil {
		return fmt.Errorf("start block file does not exist: %v, %v", sb, err)
	}
	if exists, err := exists(eb); !exists || err != nil {
		return fmt.Errorf("end block file does not exist: %v, %v", eb, err)
	}

	// test params file
	tp := filepath.Join(logDir, testParamsFile)
	if exists, err := exists(tp); !exists || err != nil {
		return fmt.Errorf("test params file does not exist: %v, %v", tp, err)
	}

	return nil
}

func exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
