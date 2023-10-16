package registry

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// Filesystem is an interface for interacting with the filesystem.
// It is used to abstract the filesystem away from the registry.
type Filesystem interface {
	// MkdirAll creates a directory and any necessary parents.
	MkdirAll(path string, perms fs.FileMode) error
	// Rename renames a file.
	Rename(oldpath, newpath string) error
	// ForEachFile calls fn for each file in the directory.
	ForEachFile(path string, fn func(name string) error) error
	// Remove removes a file.
	Remove(path string) error
}

type defaultFilesystem struct{}

// ForEachFile calls fn for each file in path.
// It passes the file's name (without the path) to fn.
func (f *defaultFilesystem) ForEachFile(path string, fn func(string) error) error {
	return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		return fn(d.Name())
	})
}

// MkDir creates a directory at path with the given permissions.
func (f *defaultFilesystem) MkdirAll(path string, perms fs.FileMode) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := checkPathLength(absolute); err != nil {
		return err
	}

	return os.MkdirAll(path, perms)
}

// Remove removes a file at path.
func (f *defaultFilesystem) Remove(path string) error {
	return os.Remove(path)
}

// Rename renames a file from oldpath to newpath.
func (f *defaultFilesystem) Rename(oldpath string, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// the below snippet is from MinIO's https://github.com/minio/minio/blob/38f35463b7fe07fbbe64bb9150d497a755c6206e/cmd/xl-storage.go#L127

// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// checkPathLength - returns error if given path name length more than 255
func checkPathLength(pathName string) error {
	// Apple OS X path length is limited to 1016
	if runtime.GOOS == "darwin" && len(pathName) > 1016 {
		return fmt.Errorf("path length exceeds limit of 1016 characters")
	}

	// On Unix we reject paths if they are just '.', '..' or '/'
	if pathName == "." || pathName == ".." || pathName == "/" {
		return fmt.Errorf("path cannot be '.' or '..' or '/'")
	}

	// Check each path segment length is > 255 on all Unix
	// platforms, look for this value as NAME_MAX in
	// /usr/include/linux/limits.h
	var count int64
	for _, p := range pathName {
		switch p {
		case '/':
			count = 0 // Reset
		default:
			count++
			if count > 255 {
				return fmt.Errorf("path segment exceeds limit of 255 characters")
			}
		}
	} // Success.
	return nil
}
