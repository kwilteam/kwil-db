package utils

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
)

func CreateDirIfNeeded(path string) error {
	return os.MkdirAll(path, 0755)
}

func ReadOrCreateFile(path string) ([]byte, error) {
	dir := filepath.Dir(path)
	if err := CreateDirIfNeeded(dir); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
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

func CreateOrOpenFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := CreateDirIfNeeded(dir); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// NOTE: os.ReadFile requires no wrapper.

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func HashFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
