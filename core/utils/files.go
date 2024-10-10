package utils

import (
	"io"
	"os"
)

// CopyFile copies a named file from src to dst. If the destination file exists,
// it is truncated.
func CopyFile(src, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	return err
}
