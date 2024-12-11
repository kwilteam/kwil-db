package utils

import "io"

// CountingWriter wraps an io.Writer, adding a Written method to get the total
// bytes written over multiple calls to Write. This is helpful if the Writer
// passes through other functions that do not return the bytes written.
type CountingWriter struct {
	w io.Writer
	c int64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{w: w}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.c += int64(n)
	return n, err
}

func (cw *CountingWriter) Written() int64 {
	return cw.c
}
