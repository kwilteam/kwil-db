package engine

import "errors"

var (
	ErrDatasetExists   = errors.New("dataset already exists")
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrDatasetNotOwned = errors.New("dataset not owned by sender")
)
