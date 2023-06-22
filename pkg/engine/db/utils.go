package db

import (
	"io"

	"github.com/kwilteam/kwil-db/pkg/sql/client"
)

// ResultsFromReader reads results from a reader (that is returned from a query) and returns them as a slice of maps.
func ResultsfromReader(r io.Reader) ([]map[string]any, error) {
	return client.ResultsfromReader(r)
}
