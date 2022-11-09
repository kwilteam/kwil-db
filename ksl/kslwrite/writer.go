package kslwrite

import (
	"bytes"
	"io"
)

type linePrefixWriter struct {
	w      io.Writer
	prefix []byte

	wrotePrefix bool
}

func newIndentWriter(w io.Writer) io.Writer {
	return &linePrefixWriter{
		w:      w,
		prefix: bytes.Repeat([]byte{' '}, 4),
	}
}

func (iw *linePrefixWriter) Write(p []byte) (n int, err error) {
	for i, b := range p {
		if b == '\n' {
			iw.wrotePrefix = false
		} else {
			if !iw.wrotePrefix {
				_, err = iw.w.Write(iw.prefix)
				if err != nil {
					return n, err
				}
				iw.wrotePrefix = true
			}
		}
		_, err = iw.w.Write(p[i : i+1])
		if err != nil {
			return n, err
		}
		n++
	}
	return len(p), nil
}
