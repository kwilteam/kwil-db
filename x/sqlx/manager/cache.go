package manager

import (
	"context"
	"fmt"
	"kwil/x/sqlx/schema"
	"sync"
	"time"
)

type CachedDB struct {
	Owner       string
	DefaultRole string
	Roles       map[string]map[string]bool // map[role][query] = true
	Queries     map[string]schema.ExecutableQuery
	Tables      map[string]schema.Table
	Indexes     map[string]schema.Index
}

type SchemaCache struct {
	DBs             map[string]*CachedDB
	MetadataManager MetadataManager
	rw              sync.RWMutex
	GCInterval      *time.Duration
}

func newCacheDB() *CachedDB {
	return &CachedDB{
		DefaultRole: "",
		Roles:       make(map[string]map[string]bool),
		Queries:     make(map[string]schema.ExecutableQuery),
		Tables:      make(map[string]schema.Table),
		Indexes:     make(map[string]schema.Index),
	}
}

// NewSchemaCache returns a new schema cache.
// intervalSeconds is the number of seconds between garbage collection runs
func NewSchemaCache(m MetadataManager, intervalSeconds int) *SchemaCache {
	gcInterval := time.Duration(intervalSeconds) * time.Second
	return &SchemaCache{
		DBs:             make(map[string]*CachedDB),
		MetadataManager: m,
		GCInterval:      &gcInterval,
	}
}

// Sync will sync a database within Postgres with the cache.  If the DB struct is not empty, it will simply cache that
func (c *SchemaCache) Sync(ctx context.Context, db *schema.Database, dbName string) error {
	if db.Owner == "" {
		// This means the database is empty and must get read from postgres
		bts, err := c.MetadataManager.GetMetadataBytes(ctx, dbName)
		if err != nil {
			return fmt.Errorf("failed to get metadata bytes: %w", err)
		}

		db = &schema.Database{}
		err = db.DecodeGOB(bts)
		if err != nil {
			return fmt.Errorf("failed to unmarshal database: %w", err)
		}
	}

	c.rw.Lock()
	defer c.rw.Unlock()

	// Setting the cache
	cacheDB := newCacheDB()
	err := cacheDB.From(db)
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.DBs[dbName] = cacheDB
	return nil
}

func (c *SchemaCache) RoleHasPermission(ctx context.Context, role string, db string, query string) (bool, error) {
	if c.DBs[db] == nil {
		return false, fmt.Errorf("db %s not found", db)
	}
	if c.DBs[db].Roles[role] == nil {
		return false, fmt.Errorf("role %s not found", role)
	}
	if c.DBs[db].Roles[role][query] {
		return true, nil
	}
	return false, nil
}

func (c *SchemaCache) WalletHasPermission(ctx context.Context, wallet string, db string, query string) (bool, error) {
	if c.DBs[db] == nil {
		return false, fmt.Errorf("db %s not found", db)
	}

	// Checking if default role has permission
	ok, err := c.RoleHasPermission(ctx, c.DBs[db].DefaultRole, db, query)
	if err != nil {
		return false, fmt.Errorf("failed to check default role: %w", err)
	}
	if ok {
		return true, nil
	}

	// Getting the roles for the wallet
	roles, err := c.MetadataManager.GetRolesByWallet(ctx, wallet, db)
	if err != nil {
		return false, fmt.Errorf("failed to get roles for wallet %s: %w", wallet, err)
	}

	// Checking if the role has permission
	for _, role := range roles {
		ok, err = c.RoleHasPermission(ctx, role, db, query)
		if err != nil {
			return false, fmt.Errorf("failed to check role %s: %w", role, err)
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}

// RunGC periodically remakes the maps in the cache since Go doesn't decrease memory allocations of maps
func (c *SchemaCache) RunGC(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(*c.GCInterval):
			c.rw.Lock()
			newDBs := make(map[string]*CachedDB)
			for dbname, cacheDB := range c.DBs {
				if cacheDB == nil {
					continue
				}

				// iterate through roles
				newRoles := make(map[string]map[string]bool)
				for role, queries := range cacheDB.Roles {
					if queries == nil {
						continue
					}

					// iterate through queries
					newQueries := make(map[string]bool)
					for query, ok := range queries {
						// I do not understand how ok could be false but I am keeping it here just in case
						if !ok {
							continue
						}

						newQueries[query] = true
					}
					newRoles[role] = newQueries
				}

				// now garbage collect the Queries
				newQueries := make(map[string]schema.ExecutableQuery)
				for queryName, executable := range cacheDB.Queries {
					// I do not have to garbage collect executable.Args since it is static
					newQueries[queryName] = executable
				}

				// tables
				newTables := make(map[string]schema.Table)
				for tableName, table := range cacheDB.Tables {
					newTables[tableName] = table
				}

				// indexes
				newIndexes := make(map[string]schema.Index)
				for indexName, index := range cacheDB.Indexes {
					newIndexes[indexName] = index
				}

				newDBs[dbname] = &CachedDB{
					DefaultRole: cacheDB.DefaultRole,
					Roles:       newRoles,
					Queries:     newQueries,
					Tables:      newTables,
					Indexes:     newIndexes,
				}
			}
			c.DBs = newDBs
			c.rw.Unlock()
		}
	}
}

func (c *SchemaCache) GetExecutable(db, query string) (*schema.ExecutableQuery, bool) {
	if c.DBs[db] == nil {
		return nil, false
	}
	exec, ok := c.DBs[db].Queries[query]
	if !ok {
		return nil, false
	}

	return &exec, true
}

func (c *CachedDB) From(db *schema.Database) error {
	// Set Default Role
	c.DefaultRole = db.DefaultRole
	c.Owner = db.Owner

	// Set Queries
	c.Queries = make(map[string]schema.ExecutableQuery)
	queries := db.Queries.GetAll()
	for queryName, query := range queries {
		executable, err := query.Prepare(db)
		if err != nil {
			return err
		}
		c.Queries[queryName] = *executable
	}

	// Set Roles
	c.Roles = make(map[string]map[string]bool)
	for roleName, role := range db.Roles {
		c.Roles[roleName] = make(map[string]bool)
		for _, query := range role.Queries {
			c.Roles[roleName][query] = true
		}
	}

	// Set Tables
	c.Tables = make(map[string]schema.Table)
	for tableName, table := range db.Tables {
		c.Tables[tableName] = table
	}

	// Set Indexes
	c.Indexes = make(map[string]schema.Index)
	for indexName, index := range db.Indexes {
		c.Indexes[indexName] = index
	}

	return nil
}

// SyncAll will sync all user-created schemas in postgres to the cache
func (c *SchemaCache) SyncAll(ctx context.Context) error {
	dbs, err := c.MetadataManager.GetAllDatabases(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all databases: %w", err)
	}

	for _, dbName := range dbs {
		db := &schema.Database{}
		err = c.Sync(ctx, db, dbName)
		if err != nil {
			fmt.Printf("failed to sync database %s: %v", dbName, err)
			fmt.Println()
			continue
		}
	}

	return nil
}

func (c *SchemaCache) Export(dbName string) (*ExportedDB, error) {
	db, ok := c.DBs[dbName]
	if !ok {
		return nil, fmt.Errorf("database %s not found", dbName)
	}

	return db.Export(dbName)
}
