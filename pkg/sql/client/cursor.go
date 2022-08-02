package client

import (
	"errors"
	"sync"
)

// This is meant to be commonly used across the application, so it is kept in its own package.

// Cursor is a cursor for iterating over results.
type Cursor struct {
	results sqlResult

	mu     sync.RWMutex
	errors []error
}

func NewCursor(results sqlResult) *Cursor {
	return &Cursor{
		results: results,
		errors:  make([]error, 0),
		mu:      sync.RWMutex{},
	}
}

func (c *Cursor) Next() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.next()
}

func (c *Cursor) next() bool {
	rowReturned, err := c.results.Next()
	if err != nil {
		c.errors = append(c.errors, err)
		return false
	}

	return rowReturned
}

func (c *Cursor) GetRecord() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.results.GetRecord()
}

// Export exports all results to a slice of maps and finishes the cursor.
// If there are any errors, they will be returned.
func (c *Cursor) Export() ([]map[string]any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	records := make([]map[string]any, 0)
	for c.next() {
		records = append(records, c.results.GetRecord())
	}

	err := c.finish()
	if err != nil {
		return nil, err
	}

	return records, nil
}

// FlushErrors returns any errors that have occurred since the last call to FlushErrors.
func (c *Cursor) FlushErrors() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.flushErrors()
}

func (c *Cursor) flushErrors() error {
	var errs []error
	for _, err := range c.errors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Finish finishes any remaining work, closes the statement, and returns any errors that have occurred.
func (c *Cursor) Finish() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.finish()
}

func (c *Cursor) finish() error {
	err := c.results.Finish()
	if err != nil {
		c.errors = append(c.errors, err)
	}

	return c.flushErrors()
}

type sqlResult interface {
	Finish() error

	Next() (bool, error)

	GetRecord() map[string]any

	Reset() error
}
