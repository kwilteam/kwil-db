/*
Package pool contains a connection pool for sql connections.

It maintains a writer thread for synchronous execution,
and a pool of reader threads for asynchronous execution.

The writer thread is "protected", meaning that the pool will never interrupt it.

The reader threads are "unprotected", meaning that the pool can interrupt them at any time.
*/
package sqlite

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/kwilteam/kwil-db/internal/sql"
	syncmap "github.com/kwilteam/kwil-db/internal/utils/sync_map"
)

var (
	ErrPoolClosed                      = errors.New("connection pool forcefully closed")
	ErrWriterInUse                     = errors.New("writer is in use")
	ErrPersistentGreaterThanMaxReaders = errors.New("persistent readers cannot be greater than max readers")
	ErrMaxReaders                      = errors.New("max readers must be greater than 0")
)

// Pool is a pool of connections.
type Pool struct {
	// opener is the function to open a connection.
	opener OpenFunc

	// closed is a channel that is closed when the pool is closed.
	closed chan struct{}

	// writerMu guards the writer.
	writerMu sync.Mutex
	// writer is the writer connection.
	writer sql.Connection

	// free is the queue of free reader connections.
	free chan sql.Connection

	// overflow is the queue of structs signaling available overflow connections.
	// overflow connections are readers that are used once and immediately closed, as opposed to being returned to the pool.
	// maximumReaders-persistentReaders = the number of overflow connections that can be open at any time.
	overflow chan struct{}

	// runningReaders is a set of cancel functions for running readers.
	runningReaders syncmap.Map[context.Context, func()]
}

// OpenFunc is a function that opens a connection.
type OpenFunc func(ctx context.Context, flags sql.ConnectionFlag) (sql.Connection, error)

// NewPool creates a new pool.
// If create is true, it will create the database if it doesn't exist.
// If create is false, it will return an error if the database doesn't exist.
func NewPool(ctx context.Context, name string, persistentReaders, maximumReaders int, create bool) (*Pool, error) {
	if persistentReaders > maximumReaders {
		return nil, fmt.Errorf(`%w: %d persistent readers, %d maximum readers`, ErrPersistentGreaterThanMaxReaders, persistentReaders, maximumReaders)
	}

	if maximumReaders < 1 {
		return nil, ErrMaxReaders
	}

	flag := sql.OpenNone
	if create {
		flag = sql.OpenCreate
	}

	writer, err := Open(ctx, name, flag)
	if err != nil {
		return nil, err
	}

	free := make(chan sql.Connection, persistentReaders)
	for i := 0; i < persistentReaders; i++ {
		conn, err := Open(ctx, name, sql.OpenReadOnly)
		if err != nil {
			writer.Close()
			return nil, err
		}

		free <- conn
	}

	overflow := make(chan struct{}, maximumReaders-persistentReaders)
	for i := 0; i < maximumReaders-persistentReaders; i++ {
		overflow <- struct{}{}
	}

	p := &Pool{
		opener: func(ctx context.Context, flags2 sql.ConnectionFlag) (sql.Connection, error) {
			return Open(ctx, name, flags2)
		},
		closed:   make(chan struct{}),
		writer:   writer,
		free:     free,
		overflow: overflow,
	}

	return p, nil
}

// isClosed returns true if the pool is closed.
func (p *Pool) isClosed() bool {
	select {
	case <-p.closed:
		return true
	default:
		return false
	}
}

// getWriter gets the writer connection.
// It assumes that the pool is already locked.
func (p *Pool) getWriter() (sql.Connection, func(), error) {
	if p.isClosed() {
		return nil, nil, ErrPoolClosed
	}

	if !p.writerMu.TryLock() {
		return nil, nil, ErrWriterInUse
	}

	returnFn := func() {
		p.writerMu.Unlock()
	}

	return p.writer, returnFn, nil
}

// Reader runs a function on a reader connection.
// it passes a context to the readers, which the pool can forcefully cancel.
func (p *Pool) reader(ctx context.Context, fn func(ctx context.Context, conn sql.Connection)) error {
	if p.isClosed() {
		return ErrPoolClosed
	}

	// we don't lock here b/c we are pulling from channels, which are thread safe
	select {
	case <-p.closed:
		return ErrPoolClosed
	case conn := <-p.free:
		defer func() {
			p.free <- conn
		}()
		// double check in case the pool was closed while we were acquiring for a free connection
		if p.isClosed() {
			return ErrPoolClosed
		}

		ctx2, cancel := context.WithCancel(ctx)
		p.runningReaders.Set(ctx2, cancel)
		defer p.runningReaders.Delete(ctx2)

		fn(ctx2, conn)

		return nil
	case <-p.overflow:
		defer func() {
			p.overflow <- struct{}{}
		}()
		// double check in case the pool was closed while we were acquiring for a free connection
		if p.isClosed() {
			return ErrPoolClosed
		}

		ctx2, cancel := context.WithCancel(ctx)
		p.runningReaders.Set(ctx2, cancel)
		defer p.runningReaders.Delete(ctx2)

		conn, err := p.opener(ctx2, sql.OpenReadOnly)
		if err != nil {
			return err
		}

		fn(ctx2, conn)

		err = conn.Close()
		if err != nil {
			return err
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close closes the pool.
func (p *Pool) Close() error {
	// close before blocking writer mutex to prevent new writers from acquiring the lock
	if p.isClosed() {
		return nil
	}
	close(p.closed)

	p.writerMu.Lock()
	defer p.writerMu.Unlock()

	var errs []error

	// cancel all running readers
	p.runningReaders.ExclusiveRead(func(m map[context.Context]func()) {
		for _, cancel := range m {
			cancel()
		}
	})

	// close all free connections
	for len(p.free) > 0 {
		conn := <-p.free

		err := conn.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	// empty the overflow channel
	// we do not close since closed channels do not block,
	// and this allows callers to open new readers
	for len(p.overflow) > 0 {
		<-p.overflow
	}

	// we have to close writer after readers
	// sqlite fails to properly clean up the WAL if the writer is closed first
	err := p.writer.Close()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

func (p *Pool) CreateSession() (sql.Session, error) {
	writer, returnFn, err := p.getWriter()
	if err != nil {
		return nil, err
	}
	defer returnFn()

	return writer.CreateSession()
}

func (p *Pool) Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error) {
	writer, returnFn, err := p.getWriter()
	if err != nil {
		return nil, err
	}
	defer returnFn()

	res, err := writer.Execute(ctx, stmt, args)
	if err != nil {
		return nil, err
	}
	defer res.Finish()

	return getResultSet(res)
}

func (p *Pool) Savepoint() (sql.Savepoint, error) {
	writer, returnFn, err := p.getWriter()
	if err != nil {
		return nil, err
	}
	defer returnFn()

	return writer.Savepoint()
}

func (p *Pool) Set(ctx context.Context, key []byte, value []byte) error {
	writer, returnFn, err := p.getWriter()
	if err != nil {
		return err
	}
	defer returnFn()

	return writer.Set(ctx, key, value)
}

func (p *Pool) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	if sync {
		writer, returnFn, err := p.getWriter()
		if err != nil {
			return nil, err
		}
		defer returnFn()

		return writer.Get(ctx, key)
	}

	var value []byte
	var queryErr error
	err := p.reader(ctx, func(ctx context.Context, conn sql.Connection) {
		value, queryErr = conn.Get(ctx, key)
	})
	if err != nil {
		return nil, err
	}
	return value, queryErr
}

func (p *Pool) Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error) {
	var res *sql.ResultSet
	var queryErr error
	err := p.reader(ctx, func(ctx context.Context, conn sql.Connection) {
		var result sql.Result
		result, queryErr = conn.Execute(ctx, query, args)
		if queryErr != nil {
			return
		}
		defer result.Finish()

		res, queryErr = getResultSet(result)
	})
	if err != nil {
		return nil, err
	}

	return res, queryErr
}

func getResultSet(res sql.Result) (*sql.ResultSet, error) {
	resultSet := &sql.ResultSet{
		ReturnedColumns: res.Columns(),
		Rows:            make([][]any, 0),
	}

	for {
		rowReturned, err := res.Next()
		if err != nil {
			return nil, err
		}
		if !rowReturned {
			break
		}

		values, err := res.Values()
		if err != nil {
			return nil, err
		}

		resultSet.Rows = append(resultSet.Rows, values)
	}

	return resultSet, nil
}
