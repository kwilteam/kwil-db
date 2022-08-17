package wal

import (
	"path"
	"sync"

	"github.com/tidwall/wal"
)

// Struct for write ahead log. Contains fields for the block that the log is for
type walWriter struct {
	mu  sync.RWMutex
	wal *wal.Log
}

//TODO: add recovery logic or a registerable recovery handler

// Will create a new WAL based on context.
func openWalWriter(dir, name string) (*walWriter, error) {
	if name != "" {
		dir = path.Join(dir, name)
	}
	innerWal, err := wal.Open(dir, nil) //need to use walOptions for reader optimization
	if err != nil {
		return nil, err
	}

	// Creating new wal
	return &walWriter{wal: innerWal}, nil
}

func openWalWriterFromHomeDir(homeDir string, name string) (*walWriter, error) {
	walDir := concatWithRootChainPath(homeDir, name)

	innerWal, err := wal.Open(walDir, nil) //need to use walOptions for reader optimization

	if err != nil {
		return nil, err
	}

	// Creating new wal
	return &walWriter{wal: innerWal}, nil
}

// Will automatically add the data to the end of the log
func (w *walWriter) appendMsgToWal(m *walMessage) error {
	if m.data == nil || len(*(m.data)) == 0 {
		return nil
	}

	return w.appendRawToWal(*m.data)
}

// Will automatically add the data to the end of the log
func (w *walWriter) appendRawToWal(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Find the last index
	currentIndex, err := w.wal.LastIndex()
	if err != nil {
		return err
	}

	// Increment by one
	currentIndex++
	// Write
	return w.wal.Write(currentIndex, data)
}

// Function to finish a wal and send it to the final directory.
func (w *walWriter) closeWal() error {
	return w.wal.Close()
}
