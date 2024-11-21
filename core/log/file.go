package log

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jrick/logrotate/rotator"
)

// This file provides functions that create io.WriteClosers suitable for use
// with WithLogWriter. Be sure to close the returned writer when done.
// To use multiple writers, just use io.MultiWriter.

// NewRotatorWriter returns a writer that writes to the specified file and
// rotates the file (zips the log to a numbered gz file and creates a new
// uncompressed log file) when the specified size is reached. If maxFiles is
// zero, the log is not rotated. Be sure to Close the returned writer.
func NewRotatorWriter(path string, maxSize int64, maxFiles int) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	return rotator.New(path, maxSize, false, maxFiles)
}

// NewFileWriter returns a writer that writes to the specified file.
func NewFileWriter(path string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}
