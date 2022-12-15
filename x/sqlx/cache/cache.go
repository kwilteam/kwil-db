package cache

import (
	"fmt"
	"kwil/x/sqlx/models"
	"sync"
)

type SchemaCache struct {
	DBs map[string]*Database
	rw  sync.RWMutex
}

// NewSchemaCache returns a new schema cache.
// intervalSeconds is the number of seconds between garbage collection runs
func New() *SchemaCache {
	return &SchemaCache{
		DBs: make(map[string]*Database),
	}
}

func (c *SchemaCache) Get(db string) *Database {
	c.rw.RLock()
	defer c.rw.RUnlock()
	return c.DBs[db]
}

func (s *SchemaCache) Store(modelDb *models.Database) error {
	s.rw.Lock()
	defer s.rw.Unlock()

	var db Database
	err := db.From(modelDb)
	if err != nil {
		return fmt.Errorf("failed to convert database: %w", err)
	}
	s.DBs[db.GetSchemaName()] = &db

	return nil
}
