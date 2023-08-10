package utils

import (
	"crypto/sha256"
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

func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
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

func WriteFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func FileStat(path string) (os.FileInfo, error) {
	return os.Stat(path)
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
