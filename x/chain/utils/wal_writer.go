package utils

import (
	"errors"
	"os"
	"path"
	"sync"

	"github.com/gofrs/flock"
	"github.com/tidwall/wal"
)

// Struct for write ahead log. Contains fields for the block that the log is for
type walWriter struct {
	mu        sync.RWMutex
	wal       *wal.Log
	lck       *flock.Flock
	lastIndex uint64
}

//TODO: add recovery logic or a registerable recovery handler

// Will create a new WAL based on context.
func openWalWriter(dir, name string) (*walWriter, error) {
	if name != "" {
		dir = path.Join(dir, name)
	}

	_ = os.MkdirAll(dir, 0755)

	lck := flock.New(dir + ".lck")
	locked, err := lck.TryLock()
	if err != nil {
		return nil, errors.New("unable to obtain wal lock: " + err.Error())
	}

	if !locked {
		return nil, errors.New("unable to obtain wal lock")
	}

	innerWal, err := wal.Open(dir, nil)
	if err != nil {
		_ = lck.Unlock()
		return nil, err
	}

	lastIndex, err := innerWal.LastIndex()
	if err != nil {
		_ = lck.Unlock()
		return nil, err
	}

	w := &walWriter{wal: innerWal, lck: lck, lastIndex: lastIndex}

	return w, nil
}

// nolint:unused
func openWalWriterFromHomeDir(homeDir string, name string) (*walWriter, error) {
	walDir := concatWithRootChainPath(homeDir, name)
	return openWalWriter(walDir, "")
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

	w.lastIndex++

	return w.wal.Write(w.lastIndex, data)
}

func (w *walWriter) closeWal() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_ = w.lck.Unlock()
	return w.wal.Close()
}

func (w *walWriter) shutdown() {
	_ = w.closeWal()
}
