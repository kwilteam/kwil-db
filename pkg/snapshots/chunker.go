package snapshots

import (
	"io"
	"os"
)

var ChunkBegin = []byte("CK_BEGIN")
var ChunkEnd = []byte("CK_END")
var BeginLen = len(ChunkBegin)
var EndLen = len(ChunkEnd)
var BoundaryLen = BeginLen + EndLen

type Chunker struct {
	reader *io.LimitedReader
	writer *io.Writer
	buf    []byte
}

func NewChunker(reader io.Reader, writer io.Writer, chunkSize int64) *Chunker {
	return &Chunker{
		reader: &io.LimitedReader{
			R: reader,
			N: chunkSize - int64(BoundaryLen),
		},
		writer: &writer,
		buf:    make([]byte, 32*1024),
	}
}

func (c *Chunker) chunkFile() error {
	err := c.beginChunk()
	if err != nil {
		return err
	}

	err = c.chunk()
	if err != nil {
		return err
	}

	err = c.endChunk()
	if err != nil {
		return err
	}
	return nil
}

func (c *Chunker) chunk() error {
	bytesWritten := uint64(0)
	for {
		n, err := c.reader.Read(c.buf)
		if err != nil {
			return err
		}
		_, err = (*c.writer).Write(c.buf[:n])
		if err != nil {
			return err
		}
		bytesWritten += uint64(n)
	}
}

func (c *Chunker) beginChunk() error {
	_, err := (*c.writer).Write(ChunkBegin)
	return err
}

func (c *Chunker) endChunk() error {
	_, err := (*c.writer).Write(ChunkEnd)
	return err
}

func copyChunkFile(srcFile string, dstFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()
	srcInfo, err := src.Stat()
	if err != nil {
		return err
	}
	ChunkSize := srcInfo.Size()

	dst, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer dst.Close()

	src.Seek(int64(BeginLen), 0)
	_, err = io.CopyN(dst, src, int64(ChunkSize)-int64(BoundaryLen))
	return err
}
