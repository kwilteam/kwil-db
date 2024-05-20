package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"github.com/kwilteam/kwil-db/core/log"
)

// StatsMonitor polls all the rpc servers in the network to collect the
// mempool stats throughout the test duration.
// Once stats monitor is stopped (ctrl-c), it retrieves all the blocks
// mined during the test duration and analyzes the block info and
// generate basic throughput metrics.

type rpcClient struct {
	// Client for the RPC server
	client *rpchttp.HTTP

	address string
}

type statsMonitor struct {
	// Client for the RPC server
	clients []*rpcClient
	// stats file name
	statsFileName string

	// Metrics
	stats *stats

	// Logger
	log log.Logger
}

type mempoolData struct {
	UnconfirmedTxs      int   `json:"unconfirmed-txs"`
	UnconfirmedTxsBytes int64 `json:"unconfirmed-tx-bytes"`
}

type blockData struct {
	Height    int64     `json:"height"`
	Time      time.Time `json:"time"`
	TxCount   int64     `json:"tx-count"`
	BlockSize int64     `json:"block-size"`
	Rounds    int32     `json:"rounds"`
	// proposer??

	// Mempool data
	UnconfirmedTxs      int   `json:"unconfirmed-txs"`
	UnconfirmedTxsBytes int64 `json:"unconfirmed-tx-bytes"`
}

type stats struct {
	// Test Boundaries
	StartBlock int64 `json:"start-block"`
	EndBlock   int64 `json:"end-block"`

	// Per Client Mempool Stats
	MempoolStats map[string]map[int64]mempoolData `json:"mempool-stats"`

	// Block data
	Blocks map[int64]blockData `json:"blocks"`

	// Metrics to be calculated
	BlockRate                float64 `json:"blockRate"`
	TransactionRate          float64 `json:"transactionRate"`
	TransactionCountPerBlock float64 `json:"transactionCountPerBlock"`
	PayloadRate              float64 `json:"payloadRate"`
	PayloadSizePerBlock      float64 `json:"payloadSizePerBlock"`

	MempoolTxRate         float64 `json:"mempoolTxRate"`
	MempoolTxSizeRate     float64 `json:"mempoolTxSizeRate"`
	MempoolTxPerBlock     float64 `json:"mempoolTxPerBlock"`
	MempoolTxSizePerBlock float64 `json:"mempoolTxSizePerBlock"`
}

func (s *stats) saveAs(filename string) error {

	bts, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bts, 0644)
}

func newStatsMonitor(addresses []string, filename string, log log.Logger) (*statsMonitor, error) {
	s := &statsMonitor{
		log:           log,
		statsFileName: filename,
		stats: &stats{
			Blocks:       make(map[int64]blockData),
			MempoolStats: make(map[string]map[int64]mempoolData),
		},
	}

	for _, address := range addresses {
		client, err := rpchttp.New(address, "/websocket")
		if err != nil {
			return nil, err
		}
		s.clients = append(s.clients, &rpcClient{
			client:  client,
			address: address,
		})
	}

	return s, nil
}

func (s *statsMonitor) Run(signalChan chan os.Signal) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.Start(ctx)

	select {
	case <-signalChan:
		s.log.Info("Received signal to stop the stats monitor")
	case <-ctx.Done():
		s.log.Info("Context is done")
		cancel()
	}

	err := s.retrieveMetrics()
	if err != nil {
		return err
	}

	s.analyze()

	err = s.stats.saveAs(s.statsFileName)
	if err != nil {
		return err
	}

	return nil
}

func (s *statsMonitor) Start(ctx context.Context) {
	s.log.Info("Starting the stats monitor")
	// Record the block height at the start of the test
	res, err := s.clients[0].client.Status(ctx)
	if err != nil {
		s.log.Error("Failed to get the chain status", log.Error(err))
		return
	}

	s.stats.StartBlock = res.SyncInfo.LatestBlockHeight

	// Start the polling routine to keep track of the unconfirmed transactions and the block height
	// for each validator node in the network (as different nodes might have different Txs in the mempool)
	for _, client := range s.clients {
		s.stats.MempoolStats[client.address] = make(map[int64]mempoolData)
		s.log.Info("Launching the polling routine for the client", log.String("address", client.address))
		go func(rc *rpcClient) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					mstats := s.stats.MempoolStats[rc.address]
					// Get current block height
					status, err := rc.client.Status(ctx)
					if err != nil {
						s.log.Error("Failed to get the chain status", log.Error(err))
						return
					}
					height := status.SyncInfo.LatestBlockHeight

					res, err := rc.client.NumUnconfirmedTxs(ctx)
					if err != nil {
						s.log.Error("Failed to get the unconfirmed transactions", log.Error(err))
						return
					}
					stats := mempoolData{
						UnconfirmedTxs:      res.Total,
						UnconfirmedTxsBytes: res.TotalBytes,
					}

					val, ok := mstats[height]
					if !ok {
						s.stats.MempoolStats[rc.address][height] = stats
					} else {
						// As we are polling every second, mempool might have accumulated more Txs
						// since the last poll so we need to update the stats. There is no way that the
						// tx count will decrease within a block height, as mempool operations are atomic
						// within block boundaries.
						if stats.UnconfirmedTxs > val.UnconfirmedTxs {
							s.stats.MempoolStats[rc.address][height] = stats
						}
					}
					time.Sleep(1 * time.Second)
				}
			}
		}(client)
	}
}

func (s *statsMonitor) retrieveMetrics() error {
	ctx := context.Background()
	s.log.Info("Retrieving the metrics")
	// Record the block height at the end of the test
	res, err := s.clients[0].client.Status(ctx)
	if err != nil {
		return err
	}
	s.stats.EndBlock = res.SyncInfo.LatestBlockHeight

	for i := s.stats.StartBlock; i <= s.stats.EndBlock; i++ {
		block, err := s.clients[0].client.Block(ctx, &i)
		if err != nil {
			return err
		}

		sz := txSize(block.Block.Txs.ToSliceOfBytes())
		val := blockData{
			Height:    i,
			Time:      block.Block.Header.Time,
			TxCount:   int64(len(block.Block.Txs)),
			BlockSize: sz,
		}

		for _, client := range s.clients {
			res, ok := s.stats.MempoolStats[client.address][i]
			if ok {
				val.UnconfirmedTxs = max(res.UnconfirmedTxs, val.UnconfirmedTxs)
				val.UnconfirmedTxsBytes = max(res.UnconfirmedTxsBytes, val.UnconfirmedTxsBytes)
			}
		}

		s.stats.Blocks[i] = val

		// Fetch and update the round number for the previous block
		round := block.Block.LastCommit.Round
		if i != s.stats.StartBlock {
			prevBlock := s.stats.Blocks[i-1]
			prevBlock.Rounds = round
			s.stats.Blocks[i-1] = prevBlock
		}
	}

	return nil
}

func (s *statsMonitor) analyze() {
	startTime := s.stats.Blocks[s.stats.StartBlock].Time
	endTime := s.stats.Blocks[s.stats.EndBlock].Time
	testDuration := endTime.Sub(startTime).Minutes()

	// Calculate the block rate
	totalBlocks := int64(len(s.stats.Blocks))
	s.stats.BlockRate = float64(totalBlocks) / testDuration

	// Calculate the transaction rate
	var totalTxs, totalPayloadSz, totalBytes, totalCount int64

	for _, block := range s.stats.Blocks {
		totalTxs += block.TxCount
		totalPayloadSz += block.BlockSize
		totalBytes += block.BlockSize
		totalCount += int64(block.UnconfirmedTxs)
	}

	s.log.Info("Analyze metrics", log.Float("test-duration(min)", testDuration), log.Int("start-block", s.stats.StartBlock), log.Int("end-block", s.stats.EndBlock), log.Int("total-blocks", totalBlocks), log.Int("total-txs", totalTxs), log.Int("total-block-size", totalPayloadSz), log.Int("total-mempool-tx-size", totalBytes), log.Int("total-mempool-txs", totalCount))

	s.stats.TransactionRate = float64(totalTxs) / testDuration
	s.stats.TransactionCountPerBlock = float64(totalTxs) / float64(totalBlocks)
	s.stats.PayloadRate = float64(totalPayloadSz) / testDuration
	s.stats.PayloadSizePerBlock = float64(totalPayloadSz) / float64(totalBlocks)
	s.stats.MempoolTxRate = float64(totalCount) / testDuration
	s.stats.MempoolTxSizeRate = float64(totalBytes) / testDuration
	s.stats.MempoolTxPerBlock = float64(totalCount) / float64(totalBlocks)
	s.stats.MempoolTxSizePerBlock = float64(totalBytes) / float64(totalBlocks)

	fmt.Println("Mempool Txs Count:", totalCount, "  Rate: ", s.stats.MempoolTxRate, " avg per block: ", s.stats.MempoolTxPerBlock)
}

func txSize(txs [][]byte) int64 {
	var size int64
	for _, tx := range txs {
		size += int64(len(tx))
	}
	return size
}
