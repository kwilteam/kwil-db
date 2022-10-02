package iox

import "io"

type multiReader struct {
	readers []io.ReadCloser
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 {
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReader); ok {
				mr.readers = r.readers
				continue
			}
		}
		n, err = mr.readers[0].Read(p)
		if err == io.EOF {
			mr.readers[0] = eofReader{}
			mr.readers = mr.readers[1:]
		}
		if n > 0 || err != io.EOF {
			if err == io.EOF && len(mr.readers) > 0 {
				err = nil
			}
			return
		}
	}
	return 0, io.EOF
}

func (mr *multiReader) Close() error {
	for _, r := range mr.readers {
		if c, ok := r.(io.Closer); ok {
			if err := c.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (mr *multiReader) WriteTo(w io.Writer) (sum int64, err error) {
	return mr.writeToWithBuffer(w, make([]byte, 1024*32))
}

func (mr *multiReader) writeToWithBuffer(w io.Writer, buf []byte) (sum int64, err error) {
	for i, r := range mr.readers {
		var n int64
		if subMr, ok := r.(*multiReader); ok {
			n, err = subMr.writeToWithBuffer(w, buf)
		} else {
			n, err = io.CopyBuffer(w, r, buf)
		}
		sum += n
		if err != nil {
			mr.readers = mr.readers[i:]
			return sum, err
		}
		mr.readers[i] = nil
	}
	mr.readers = nil
	return sum, nil
}

var _ io.WriterTo = (*multiReader)(nil)

// MultiReadCloser returns a io.ReadCloser that's the logical concatenation of
// the provided input readers. They're read sequentially. Once all
// inputs have returned EOF, Read will return EOF.  If any of the readers
// return a non-nil, non-EOF error, Read will return that error.
func MultiReadCloser(readers ...io.ReadCloser) io.ReadCloser {
	r := make([]io.ReadCloser, len(readers))
	copy(r, readers)
	return &multiReader{r}
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (eofReader) Close() error {
	return nil
}
