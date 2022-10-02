package schema

import (
	"fmt"
	"io"
	"kwil/x/iox"
	"os"
	"path/filepath"
)

func GatherFiles(paths ...string) (rd io.ReadCloser, err error) {
	var files []io.ReadCloser
	defer func() {
		if err != nil {
			for _, f := range files {
				f.Close()
			}
		}
	}()

	for _, path := range paths {
		switch stat, err := os.Stat(path); {
		case err != nil:
			continue
		case stat.IsDir():
			dir, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			for _, f := range dir {
				if f.IsDir() {
					continue
				}
				if f, ok := mayOpen(filepath.Join(path, f.Name())); ok {
					files = append(files, f)
				}
			}
		default:
			f, ok := mayOpen(path)
			if !ok {
				return nil, fmt.Errorf("invalid file: %s", path)
			}
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no schema files found in: %s", paths)
	}

	return iox.MultiReadCloser(files...), nil
}

func mayOpen(path string) (io.ReadCloser, bool) {
	if filepath.Ext(path) != ".hcl" {
		return nil, false
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	return f, true
}
