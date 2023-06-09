package utils

import (
	"io"
	"os"
	"path/filepath"
)

func CreateDirIfNeeded(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, os.ModePerm)
}

func ReadOrCreateFile(path string, permissions int) ([]byte, error) {
	if err := CreateDirIfNeeded(path); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, permissions, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func CreateOrOpenFile(path string, permissions int) (*os.File, error) {
	if err := CreateDirIfNeeded(path); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, permissions, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func OpenFile(path string, permissions int) (*os.File, error) {
	file, err := os.OpenFile(path, permissions, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}
