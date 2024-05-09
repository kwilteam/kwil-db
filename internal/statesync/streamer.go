package statesync

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/core/log"
)

// Utility to stream chunks of a snapshot
type Streamer struct {
	log               log.Logger
	numChunks         uint32
	files             []string
	currentChunk      *os.File
	currentChunkIndex uint32
}

func NewStreamer(numChunks uint32, chunkDir string, logger log.Logger) *Streamer {
	files := make([]string, numChunks)
	for i := uint32(0); i < numChunks; i++ {
		file := filepath.Join(chunkDir, fmt.Sprintf("chunk-%d.sql.gz", i))
		files[i] = file
	}

	return &Streamer{
		log:       logger,
		numChunks: numChunks,
		files:     files,
	}
}

// Next opens the next chunk file for streaming
func (s *Streamer) Next() error {
	if s.currentChunk != nil {
		s.currentChunk.Close()
	}

	if s.currentChunkIndex >= s.numChunks {
		return io.EOF // no more chunks to stream
	}

	file, err := os.Open(s.files[s.currentChunkIndex])
	if err != nil {
		return fmt.Errorf("failed to open chunk file: %w", err)
	}

	s.currentChunk = file
	s.currentChunkIndex++

	return nil
}

func (s *Streamer) Close() error {
	if s.currentChunk != nil {
		s.currentChunk.Close()
	}

	return nil
}

// Read reads from the current chunk file
// If the current chunk is exhausted, it opens the next chunk file
// until all chunks are read
func (s *Streamer) Read(p []byte) (n int, err error) {
	if s.currentChunk == nil {
		if err := s.Next(); err != nil {
			return 0, err
		}
	}

	n, err = s.currentChunk.Read(p)
	if err == io.EOF {
		err = s.currentChunk.Close()
		s.currentChunk = nil
		if s.currentChunkIndex < s.numChunks {
			return s.Read(p)
		}
	}
	return n, err
}
