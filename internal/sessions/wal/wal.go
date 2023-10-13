package wal

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/tidwall/wal"
)

const (
	startingWriteIdx = 1
	startingReadIdx  = 1
)

// Wal is a basic write-ahead log.
// This can and should be improved on.
// At the time of its creation, this is
// only made to implement sessions.Wal.
type Wal struct {
	path string

	mu sync.Mutex

	// tidwall/wal starts indexing at 1.
	wal *wal.Log

	writeIdx uint64
	readIdx  uint64
}

// OpenWal opens a wal at the given path.
// If the wal does not exist, it will be created.
// For example, passing in "/tmp/wal" will create
// a wal at "/tmp/wal.wal".
func OpenWal(path string) (*Wal, error) {
	w, err := wal.Open(path, nil)
	if err != nil {
		return nil, err
	}

	return &Wal{
		path:     path,
		wal:      w,
		writeIdx: startingWriteIdx,
		readIdx:  startingReadIdx,
	}, nil
}

// Append appends a new entry to the WAL
func (w *Wal) Append(ctx context.Context, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.wal.Write(w.writeIdx, data)
	if err != nil {
		return err
	}

	w.writeIdx++

	return nil
}

// ReadNext reads the next entry from the WAL
// it will return an io.EOF when it has reached the end of the WAL
// it will also return an io.EOF if the wal is corrupt
func (w *Wal) ReadNext(ctx context.Context) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	data, err := w.wal.Read(w.readIdx)
	if err == wal.ErrNotFound || err == wal.ErrCorrupt {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	w.readIdx++

	return data, nil
}

// Truncate deletes the entire WAL
// and opens a new one.
func (w *Wal) Truncate(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.wal.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(w.path)
	if err != nil {
		return err
	}

	w.wal, err = wal.Open(w.path, nil)
	if err != nil {
		return err
	}

	w.writeIdx = startingWriteIdx
	w.readIdx = startingReadIdx

	return nil
}

// Close closes the WAL
func (w *Wal) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.wal.Close()
}
