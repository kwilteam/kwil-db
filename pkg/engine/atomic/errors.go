package atomic

import "errors"

var (
	ErrDatasetExists   = errors.New("dataset already exists")
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrOpeningDataset  = errors.New("failed to open dataset database file")
	ErrTableExists     = errors.New("table already exists")
)
