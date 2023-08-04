package utils

import (
	"bufio"
	"os"
	"sync"
)

type Wal struct {
	Path        string
	Wal         *os.File
	mu          sync.Mutex
	maxLineSize int
}

func NewWal(path string) (*Wal, error) {

	wal, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Wal{
		Path:        path,
		Wal:         wal,
		maxLineSize: 0,
	}, nil
}

func (w *Wal) Write(data []byte) error {
	_, err := w.Wal.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) Close() error {
	err := w.Wal.Close()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) Truncate() error {
	err := w.Wal.Truncate(0)
	if err != nil {
		return err
	}
	err = w.Wal.Sync()
	if err != nil {
		return err
	}
	w.maxLineSize = 0
	return nil
}

func (w *Wal) Sync() error {
	err := w.Wal.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) WriteSync(data []byte) error {
	_, err := w.Wal.Write(data)
	if err != nil {
		return err
	}
	err = w.Wal.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) OverwriteSync(data []byte) error {
	err := w.Wal.Truncate(0)
	if err != nil {
		return err
	}
	_, err = w.Wal.Write(data)
	if err != nil {
		return err
	}
	err = w.Wal.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) UpdateMaxLineSz(sz int) {
	if sz > w.maxLineSize {
		w.maxLineSize = sz
	}
}

func (w *Wal) Lock() {
	w.mu.Lock()
}

func (w *Wal) Unlock() {
	w.mu.Unlock()
}

func (w *Wal) NewScanner() *bufio.Scanner {
	w.Wal.Seek(0, 0)
	scanner := bufio.NewScanner(w.Wal)
	buf := make([]byte, w.maxLineSize)
	scanner.Buffer(buf, w.maxLineSize)
	scanner.Split(bufio.ScanLines)
	return scanner
}

func (w *Wal) Read() []byte {
	bts, err := os.ReadFile(w.Path)
	if err != nil {
		return nil
	}
	return bts
}

func (w *Wal) IsEmpty() bool {
	stat, err := w.Wal.Stat()
	if err != nil {
		return true
	}
	if stat.Size() == 0 {
		return true
	}
	return false
}
