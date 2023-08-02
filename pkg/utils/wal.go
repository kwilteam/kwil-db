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
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.Wal.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.Wal.Close()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()
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

func (w *Wal) TruncateUnSafe() error {
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
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.Wal.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wal) WriteSync(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
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
	w.mu.Lock()
	defer w.mu.Unlock()
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
	w.mu.Lock()
	defer w.mu.Unlock()
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
	scanner := bufio.NewScanner(w.Wal)
	buf := make([]byte, w.maxLineSize)
	scanner.Buffer(buf, w.maxLineSize)
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
	w.mu.Lock()
	defer w.mu.Unlock()
	stat, err := w.Wal.Stat()
	if err != nil {
		return true
	}
	if stat.Size() == 0 {
		return true
	}
	return false
}
