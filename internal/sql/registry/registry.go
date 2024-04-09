/*
	Package registry is responsible for transactionally committing to databases.

	It can create and delete files, as well as execute DML.

	It uses an idempotent key within each database to track the changes that have been made to the database.  On crash recovery,
	it uses this key to determine whether or not the database has been committed to.

	When a database is created, it is created in a temporary file.  When the commit is called, it is renamed to the correct file.
	When a database is deleted, it is renamed to a deleted file.  When the commit is called, it is deleted.

	When a database is opened, a savepoint is opened with it.  This savepoint is rolled back on rollback, and committed on commit.
	When a database is opened, it is opened with a writer connection.  This connection is returned on rollback, and closed on commit.

	If a database already contains the idempotent key, it will return nil for any incoming operation.
*/

package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kwilteam/kwil-db/core/log"
	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/utils/order"
	syncmap "github.com/kwilteam/kwil-db/internal/utils/sync_map"
	"go.uber.org/zap"
)

// Registry is used to register databases.
// It manages connections, as well as atomicity.
type Registry struct {
	// mu is the mutex for the registry.
	mu sync.RWMutex

	log log.Logger

	// opener is the function to open a database.
	opener PoolOpener

	// directory is the directory where the databases are stored.
	directory string

	// pools is a map of database ids to their connection pools.
	// we use a sync map since pools is accessed in `Query`, which
	// is not protected by the mutex.
	pools syncmap.Map[string, Pool]

	// session is the current session.
	session *session

	// readerCloseTime is the time the engine will wait before forcibly closing a reader when committing.
	readerCloseTime time.Duration

	// filesystem is the filesystem to use.
	filesystem Filesystem

	// openReaderChan is a channel that controls when readers can be opened
	openReaderChan chan struct{}
}

// NewRegistry opens a registry.
// It does not handle recovery, and it is up to the caller to determine if recovery is needed.
// If there was a failure in between `Begin` and `Commit`, then NewRegistry will reset the state
// of the registry to the state before `Begin` was called.
func NewRegistry(ctx context.Context, opener PoolOpener, directory string, opts ...RegistryOpt) (*Registry, error) {
	readerChan := make(chan struct{}, 1)
	readerChan <- struct{}{}

	r := &Registry{
		opener:          opener,
		directory:       directory,
		log:             log.NewNoOp(),
		readerCloseTime: time.Duration(100) * time.Millisecond,
		openReaderChan:  readerChan,
		filesystem:      &defaultFilesystem{},
	}

	for _, opt := range opts {
		opt(r)
	}

	fmt.Println("LOOK FOR ME:" + fmt.Sprint(r.readerCloseTime))

	// create the directory if it does not exist
	err := r.filesystem.MkdirAll(directory, 0755)
	if err != nil {
		return nil, err
	}

	err = r.cleanupInFlightDBs(ctx)
	if err != nil {
		return nil, err
	}

	dbs, err := r.listFromFiles(ctx)
	if err != nil {
		return nil, err
	}

	for _, dbid := range dbs {
		pool, err := r.opener(ctx, r.path(dbid), false)
		if err != nil {
			return nil, err
		}
		r.pools.Set(dbid, pool)
	}
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Create creates a new database.
// If the database already exists, it returns an error.
func (r *Registry) Create(ctx context.Context, dbid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		return ErrRegistryNotWritable
	}

	// if in recovery, the database might have already been created
	if r.session.Recovery {
		_, ok := r.pools.Get(dbid)
		if ok {
			return nil
		}
	}

	// check if the database already exists
	_, ok := r.pools.Get(dbid)
	if ok {
		return ErrDatabaseExists
	}

	// create the database
	pool, err := r.opener(ctx, r.path(dbid)+newSuffix, true)
	if err != nil {
		return err
	}

	sp, err := pool.Savepoint()
	if err != nil {
		return err
	}

	session, err := pool.CreateSession()
	if err != nil {
		sp.Rollback()
		return err
	}

	r.pools.Set(dbid, pool)

	r.session.Open[dbid] = &openDB{
		Pool:      pool,
		Savepoint: sp,
		Session:   session,
		Status:    dbStatusNew,
	}

	return nil
}

// Delete deletes a database.
// If the database does not exist, it returns an error.
// If the database was created within the same session, it returns an error.
func (r *Registry) Delete(ctx context.Context, dbid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		return ErrRegistryNotWritable
	}

	// if in recovery, the database might have already been deleted
	if r.session.Recovery {
		_, ok := r.pools.Get(dbid)
		if !ok {
			return nil
		}
	}

	// check if the database exists
	pool, ok := r.pools.Get(dbid)
	if !ok {
		return ErrDatabaseNotFound
	}

	// by default, the file is called path/dbid
	// if it is not committed, then it is called path/dbid.new
	fileName := dbid

	// check if the database was created in this session
	openDb, ok := r.session.Open[dbid]
	if ok {

		// if already used in this session, make sure it was not created or deleted
		if openDb.Status == dbStatusNew {
			fileName += newSuffix
		}

		err := openDb.Savepoint.Rollback()
		if err != nil {
			return err
		}

		err = openDb.Session.Delete()
		if err != nil {
			return err
		}

		delete(r.session.Open, dbid)
	}

	err := pool.Close()
	if err != nil {
		return err
	}

	err = r.filesystem.Rename(r.path(fileName), r.path(fileName)+deletedSuffix)
	if err != nil {
		return err
	}

	r.pools.Delete(dbid)

	return nil
}

// Execute executes a statement on a database.
// If the database does not exist, it returns an error.
func (r *Registry) Execute(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		return nil, ErrRegistryNotWritable
	}

	conn, skip, err := r.getExistentWriter(ctx, dbid, r.session.IdempotencyKey)
	if err != nil {
		return nil, err
	}

	if skip {
		return &sql.ResultSet{
			ReturnedColumns: []string{},
			Rows:            [][]any{},
		}, nil
	}

	// execute the statement
	// we do not return the connection, as it will be returned on rollback or closed on commit
	return conn.Execute(ctx, stmt, params)
}

// Set sets the value for a key in a database's key value store.
func (r *Registry) Set(ctx context.Context, dbid string, key, value []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// check if writable, since r.session will be nil if not writable
	if r.session == nil {
		return ErrRegistryNotWritable
	}

	conn, skip, err := r.getExistentWriter(ctx, dbid, r.session.IdempotencyKey)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	return conn.Set(ctx, key, value)
}

// getExistentWriter gets a writer for a database that already exists.
// If the connection should be skipped due to the idempotency key, it returns shouldSkip as true.
func (r *Registry) getExistentWriter(ctx context.Context, dbid string, idempotencyKey []byte) (conn Pool, shouldSkip bool, err error) {
	// if recovery, check if the idempotency key matches
	if r.session.Recovery {
		pool, ok := r.pools.Get(dbid)
		if !ok {
			return nil, false, ErrDatabaseNotFound
		}

		openDb, ok := r.session.Open[dbid]
		if ok {
			conn = openDb.Pool
		} else {
			var err error
			conn = pool

			sp, err := conn.Savepoint()
			if err != nil {
				return nil, false, err
			}

			session, err := conn.CreateSession()
			if err != nil {
				sp.Rollback()
				return nil, false, err
			}

			r.session.Open[dbid] = &openDB{
				Pool:      conn,
				Savepoint: sp,
				Session:   session,
				Status:    dbStatusExists,
			}
		}

		idempotencyKey, err := getIdempotencyKey(ctx, conn)
		if err != nil {
			return nil, false, err
		}

		// if the idempotency key matches, this DB has already been executed
		if bytes.Equal(idempotencyKey, r.session.IdempotencyKey) {
			return conn, true, nil
		}
	}

	// check if the database exists
	pool, ok := r.pools.Get(dbid)
	if !ok {
		return nil, false, ErrDatabaseNotFound
	}

	// get the writer, either from the session or from the pool
	openDb, ok := r.session.Open[dbid]
	if ok {
		conn = openDb.Pool
	} else {
		// if not yet used, open it and register it
		var err error
		conn = pool

		sp, err := conn.Savepoint()
		if err != nil {
			return nil, false, err
		}

		session, err := conn.CreateSession()
		if err != nil {
			sp.Rollback()
			return nil, false, err
		}

		r.session.Open[dbid] = &openDB{
			Pool:      conn,
			Savepoint: sp,
			Session:   session,
			Status:    dbStatusExists,
		}
	}

	return conn, false, nil
}

// Query executes a query on a database.
// If the database does not exist, it returns an error.
func (r *Registry) Query(ctx context.Context, dbid string, stmt string, params map[string]any) (*sql.ResultSet, error) {
	pool, ok := r.pools.Get(dbid)
	if !ok {
		return nil, ErrDatabaseNotFound
	}

	// get the reader context
	ctx2, cancel, err := r.getReaderCtx(ctx, dbid)
	if err != nil {
		return nil, err
	}
	defer cancel()

	// execute the statement
	return pool.Query(ctx2, stmt, params)
}

// Get gets the value for a key in a database's key value store.
func (r *Registry) Get(ctx context.Context, dbid string, key []byte, sync bool) ([]byte, error) {

	if sync {
		r.mu.Lock()
		defer r.mu.Unlock()

		// check if writable, since r.session will be nil if not writable
		if r.session == nil {
			return nil, ErrRegistryNotWritable
		}

		conn, skip, err := r.getExistentWriter(ctx, dbid, r.session.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if skip {
			return nil, nil
		}

		return conn.Get(ctx, key, true)
	}

	// check if the database exists
	pool, ok := r.pools.Get(dbid)
	if !ok {
		return nil, ErrDatabaseNotFound
	}

	ctx2, cancel, err := r.getReaderCtx(ctx, dbid)
	if err != nil {
		return nil, err
	}
	defer cancel()

	// execute the statement
	return pool.Get(ctx2, key, false)
}

// getReaderCtx gets a context for a reader.
// it sets a timeout to prevent readers from being open for too long.
func (r *Registry) getReaderCtx(ctx context.Context, dbid string) (context.Context, func(), error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.session != nil {
		openDb, ok := r.session.Open[dbid]
		if ok && (openDb.Status == dbStatusNew || openDb.Status == dbStatusDeleted) {
			return nil, nil, ErrDatabaseNotFound
		}
	}

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case <-r.openReaderChan:
		r.openReaderChan <- struct{}{}
	}

	ctx2, cancel := context.WithTimeout(ctx, r.readerCloseTime)

	return ctx2, cancel, nil
}

// blockReaders cancels all readers.
// it will wait for the configured time before forcibly closing them.
// it returns a function that unblocks all readers.
func (r *Registry) blockReaders() (unblock func()) {
	<-r.openReaderChan
	unblock = func() {
		r.openReaderChan <- struct{}{}
	}

	return unblock
}

// List lists all databases.
func (r *Registry) List(ctx context.Context) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var dbs []string

	r.pools.ExclusiveRead(func(m map[string]Pool) {
		for dbid := range m {
			dbs = append(dbs, dbid)
		}
	})

	return dbs, nil
}

// listFromFiles lists all databases from the files.
// this is very expensive, and should only be used during startup.
func (r *Registry) listFromFiles(ctx context.Context) ([]string, error) {
	var dbs []string

	err := r.forEach(ctx, regexp.MustCompile(``), func(fileName string) error {
		dbs = append(dbs, fileName)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dbs, nil
}

// Close closes the registry.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error

	if r.session != nil {
		for _, openDb := range r.session.Open {
			err := openDb.Savepoint.Rollback()
			if err != nil {
				errs = append(errs, err)
			}

			err = openDb.Session.Delete()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	r.pools.ExclusiveRead(func(m map[string]Pool) {
		for _, pool := range m {
			err := pool.Close()
			if err != nil {
				errs = append(errs, err)
			}
		}
	})

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// Begin signals the start of a session.
// A session is simply a series of operations across many databases that are executed atomically.
func (r *Registry) Begin(ctx context.Context, idempotencyKey []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// if already in a session, return an error
	if r.session != nil {
		return ErrAlreadyInSession
	}

	if r.session != nil {
		return ErrWritable
	}

	err := r.cleanupInFlightDBs(ctx)
	if err != nil {
		return err
	}

	r.session = newSession(idempotencyKey, false)

	return nil
}

// BeginRecovery signals the start of a session recovery.
// Session recovery occurs when a session is is interrupted during the commit.
// It is the caller's responsibility to identify that a session recovery is needed.
// Recovery is only needed if the session was interrupted during `Commit`.
// If a failure occurs between `Begin` and `Commit`, then recovery is not needed.
func (r *Registry) BeginRecovery(ctx context.Context, idempotencyKey []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// if already in a session, return an error
	if r.session != nil {
		return ErrAlreadyInSession
	}

	if r.session != nil {
		return ErrWritable
	}

	err := r.cleanupInFlightDBs(ctx)
	if err != nil {
		return err
	}

	r.session = newSession(idempotencyKey, true)

	return nil
}

// Commit signals the end of a session.
// All databases will be committed, in order.
func (r *Registry) Commit(ctx context.Context, idempotencyKey []byte) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		return nil, ErrRegistryNotWritable
	}

	// if the idempotency key does not match, return an error
	if !bytes.Equal(r.session.IdempotencyKey, idempotencyKey) {
		return nil, fmt.Errorf("%w: expected %x, got %x", ErrIdempotencyKeyMismatch, r.session.IdempotencyKey, idempotencyKey)
	}

	// cancel all readers
	unblock := r.blockReaders()
	defer unblock()

	// maps dbid to app hash
	appHashes := make(map[string][]byte)

	var errs []error
	r.pools.Exclusive(func(pools map[string]Pool) {
		for dbid, pool := range pools {
			openDb, ok := r.session.Open[dbid]
			if !ok {
				// if not ok, the database has been skipped this session
				// we should check if it has the idempotency key
				// if so, then we need to include its app hash in the commit

				// get the idempotency key
				idempotencyKey, err := getIdempotencyKey(ctx, pool)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				if !bytes.Equal(idempotencyKey, r.session.IdempotencyKey) {
					// if the idempotency key does not match, we can skip this database
					continue
				}

				// get the app hash
				appHash, err := getAppHash(ctx, pool)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				appHashes[dbid] = appHash
				r.log.Debug("adding app hash", zap.String("dbid", dbid), zap.String("appHash", fmt.Sprintf("%x", appHash)), zap.Binary("idempotencyKey", idempotencyKey))

				continue
			}

			switch openDb.Status {
			default:
				panic("invalid status")
			case dbStatusNew:
				appHash, err := openDb.Session.ChangesetID(ctx)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = setAppHash(ctx, openDb.Pool, appHash)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = setIdempotencyKey(ctx, openDb.Pool, r.session.IdempotencyKey)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = openDb.Savepoint.Commit()
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = openDb.Session.Delete()
				if err != nil {
					errs = append(errs, err)
					continue
				}

				// close the pool to allow the rename
				err = pool.Close()
				if err != nil {
					errs = append(errs, err)
					r.log.Debug("failed to close pool", zap.String("dbid", dbid), zap.Error(err))
					continue
				}

				err = r.filesystem.Rename(r.path(dbid)+newSuffix, r.path(dbid))
				if err != nil {
					errs = append(errs, err)
					r.log.Debug("failed to rename db file", zap.String("from", dbid+newSuffix), zap.String("to", dbid), zap.Error(err))
					continue
				}

				r.log.Debug("renamed db file", zap.String("from", dbid+newSuffix), zap.String("to", dbid))

				delete(r.session.Open, dbid)
				appHashes[dbid] = appHash

				r.log.Debug("adding app hash", zap.String("dbid", dbid), zap.String("appHash", fmt.Sprintf("%x", appHash)), zap.Binary("idempotencyKey", idempotencyKey))

				newPool, err := r.opener(ctx, r.path(dbid), false)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				pools[dbid] = newPool
			case dbStatusExists:
				appHash, err := openDb.Session.ChangesetID(ctx)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = setAppHash(ctx, openDb.Pool, appHash)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = setIdempotencyKey(ctx, openDb.Pool, r.session.IdempotencyKey)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = openDb.Savepoint.Commit()
				if err != nil {
					errs = append(errs, err)
					continue
				}

				err = openDb.Session.Delete()
				if err != nil {
					errs = append(errs, err)
					continue
				}

				delete(r.session.Open, dbid)

				appHashes[dbid] = appHash
				r.log.Debug("adding app hash", zap.String("dbid", dbid), zap.String("appHash", fmt.Sprintf("%x", appHash)), zap.Binary("idempotencyKey", idempotencyKey))

			}
		}
	})
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// remove deleted databases
	err := r.forEach(ctx, deletedSuffixRegex, func(fileName string) error {
		dbid := removeDeletedSuffix(fileName)

		// if deleted, delete
		err := r.filesystem.Remove(r.path(dbid) + deletedSuffix)
		if err != nil {
			return err
		}

		r.session.Committed[dbid] = []byte{}

		return nil
	})
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	for _, committed := range order.OrderMap(appHashes) {
		_, err := hasher.Write(committed.Value)
		if err != nil {
			return nil, err
		}
	}

	r.session = nil

	return hasher.Sum(nil), nil
}

// Cancel signals that the session should be cancelled.
// If no session is in progress, it returns nil.
func (r *Registry) Cancel(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		return nil
	}

	unblock := r.blockReaders()
	defer unblock()

	clear(r.session.Committed)

	var errs []error
	for dbid, openDb := range r.session.Open {
		switch openDb.Status {
		case dbStatusNew:
			err := openDb.Savepoint.Rollback()
			if err != nil {
				errs = append(errs, err)
			}

			err = r.filesystem.Remove(r.path(dbid) + newSuffix)
			if err != nil {
				errs = append(errs, err)
			}

			r.pools.Delete(dbid)
		case dbStatusDeleted:
			err := r.filesystem.Rename(r.path(dbid)+deletedSuffix, r.path(dbid))
			if err != nil {
				errs = append(errs, err)
			}

			newPool, err := r.opener(ctx, r.path(dbid), false)
			if err != nil {
				errs = append(errs, err)
			}

			r.pools.Set(dbid, newPool)
		case dbStatusExists:
			err := openDb.Savepoint.Rollback()
			if err != nil {
				errs = append(errs, err)
			}
		default:
			panic("invalid status")
		}
	}

	clear(r.session.Open)
	r.session = nil

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// cleanupInFlightDBs cleans up all databases in an uncommitted deploy or drop.
// It will delete uncommitted new ones, or undelete uncommitted deleted ones.
func (r *Registry) cleanupInFlightDBs(ctx context.Context) error {
	// cleanup deleted databases by renaming them to their original names
	if err := r.forEach(ctx, deletedSuffixRegex, func(fileName string) error {
		r.log.Debug("recovering deleted database", zap.String("dbid", removeDeletedSuffix(fileName)))

		err := r.filesystem.Rename(r.path(fileName), r.path(removeDeletedSuffix(fileName)))
		if err != nil {
			return err
		}

		dbid := removeDeletedSuffix(fileName)

		pool, err := r.opener(ctx, r.path(dbid), false)
		if err != nil {
			return err
		}

		r.pools.Set(dbid, pool)

		return nil
	}); err != nil {
		return err
	}

	// cleanup new databases by deleting them
	if err := r.forEach(ctx, newSuffixRegex, func(fileName string) error {
		r.log.Debug("reverting creation of new database", zap.String("dbid", fileName))

		r.pools.Delete(fileName)
		return r.filesystem.Remove(r.path(fileName))
	}); err != nil {
		return err
	}

	return nil
}

// forEach iterates over all files in the registry directory that match the regex.
// It calls the function for each file.
// it will ignore -shm and -wal files.
func (r *Registry) forEach(ctx context.Context, regex *regexp.Regexp, fn func(fileName string) error) error {
	return r.filesystem.ForEachFile(r.directory, func(name string) error {
		if regex.MatchString(name) && !strings.HasSuffix(name, "-shm") && !strings.HasSuffix(name, "-wal") {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return fn(name)
			}
		}
		return nil
	})
}

// path returns the path to the database.
func (r *Registry) path(dbid string) string {
	return filepath.Join(r.directory, dbid)
}

const (
	// deletedSuffix is the suffix for deleted databases.
	deletedSuffix = ".deleted"
	// newSuffix is the suffix for new databases.
	newSuffix = ".new"
)

// newSuffixRegex is the regex for new databases.
// we intentionally do not enforce that the suffix is not at the end here, b/c we can
// have a file called dbid.new.deleted, which is a deleted database that was never committed.
var newSuffixRegex = regexp.MustCompile(`\.new`)

// deletedSuffixRegex is the regex for deleted databases.
var deletedSuffixRegex = regexp.MustCompile(`\.deleted$`)

// removeDeletedSuffix removes the deleted suffix from a database name.
func removeDeletedSuffix(s string) string {
	return s[:len(s)-len(deletedSuffix)]
}

// getIdempotencyKey gets the most recently persisted idempotency key for a database.
func getIdempotencyKey(ctx context.Context, conn KV) ([]byte, error) {
	return conn.Get(ctx, idempotencyKeyKey, true)
}

// setIdempotencyKey sets the idempotency key for a database.
func setIdempotencyKey(ctx context.Context, conn KV, idempotencyKey []byte) error {
	return conn.Set(ctx, idempotencyKeyKey, idempotencyKey)
}

// getAppHash gets the most recently persisted app hash for a database.
func getAppHash(ctx context.Context, conn KV) ([]byte, error) {
	return conn.Get(ctx, appHashKey, true)
}

// setAppHash sets the app hash for a database.
func setAppHash(ctx context.Context, conn KV, appHash []byte) error {
	return conn.Set(ctx, appHashKey, appHash)
}

var (
	// idempotencyKeyKey is the key for the idempotency key.
	idempotencyKeyKey = []byte("idempotency_key")
	// appHashKey is the key for the app hash.
	appHashKey = []byte("app_hash")
)
