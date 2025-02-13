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

// CountingReader wraps an io.Reader, adding a ReadCount method to get the total
// bytes read over multiple calls to Read. This is helpful if the Reader passes
// through other functions that do not return the bytes read.
type CountingReader struct {
	r io.Reader
	c int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{r: r}
}

var _ io.Reader = (*CountingReader)(nil)

func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.c += int64(n)
	return n, err
}

var _ io.ByteReader = (*CountingReader)(nil)

func (cr *CountingReader) ReadByte() (byte, error) {
	var b [1]byte
	_, err := cr.Read(b[:]) // must call our Read to count this byte
	return b[0], err
}

func (cr *CountingReader) ReadCount() int64 {
	return cr.c
}
