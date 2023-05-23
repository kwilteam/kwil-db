package dto

// DatasetContext is a context for a dataset.
// Once provided, it should not be modified.
type DatasetContext struct {
	// Name is the name of the dataset.
	Name string
	// Owner is the owner of the dataset.
	Owner string
}
