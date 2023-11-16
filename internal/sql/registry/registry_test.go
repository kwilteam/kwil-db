package registry_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"io/fs"

	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/kwilteam/kwil-db/internal/sql/registry"
	"github.com/stretchr/testify/require"
)

// Testing regular registry operations.
func Test_Registry(t *testing.T) {
	type testCase struct {
		name          string
		initialDbs    []string
		fn            func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) // special recovery flag to handle special case with kv Get
		hasPools      []string
		hasExecutions map[string][]executedStmt
	}

	tests := []testCase{
		{
			name: "create database",
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				err := registry.Create(ctx, "db1")
				require.NoError(t, err)
			},
			hasPools: []string{"db1"},
		},
		{
			name: "create database and execute some statements",
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				err := registry.Create(ctx, "db1")
				require.NoError(t, err)

				_, err = registry.Execute(ctx, "db1", "CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT)", nil)
				require.NoError(t, err)
			},
			hasPools: []string{"db1"},
			hasExecutions: map[string][]executedStmt{
				"db1": {
					{
						stmt:   "CREATE TABLE foo (id INTEGER PRIMARY KEY, name TEXT)",
						params: nil,
					},
				},
			},
		},
		{
			name:       "execute against existing database",
			initialDbs: []string{"db1"},
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				_, err := registry.Execute(ctx, "db1", "INSERT", nil)
				require.NoError(t, err)
			},
			hasPools: []string{"db1"},
			hasExecutions: map[string][]executedStmt{
				"db1": {
					{
						stmt:   "INSERT",
						params: nil,
					},
				},
			},
		},
		{
			name:       "delete database",
			initialDbs: []string{"db1"},
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				err := registry.Delete(ctx, "db1")
				require.NoError(t, err)
			},
			hasPools: []string{},
		},
		{
			name:       "uncommitted dbs are removed when the registry opens",
			initialDbs: []string{"db1.new", "db2.deleted"},
			fn:         func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {},
			hasPools:   []string{},
		},
		{
			name:       "adding data after deleting a database fails",
			initialDbs: []string{"db1"},
			fn: func(ctx context.Context, t *testing.T, r *registry.Registry, recovery bool) {
				err := r.Delete(ctx, "db1")
				require.NoError(t, err)

				res, err := r.Execute(ctx, "db1", "INSERT", nil)
				require.Equal(t, registry.ErrDatabaseNotFound, err)
				require.Nil(t, res)
			},
			hasPools: []string{},
		},
		{
			name:       "testing kv",
			initialDbs: []string{"db1"},
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				err := registry.Set(ctx, "db1", []byte("key"), []byte("value"))
				require.NoError(t, err)

				if recovery {
					// edge case with getting on recovery
					return
				}

				val, err := registry.Get(ctx, "db1", []byte("key"), true)
				require.NoError(t, err)

				require.Equal(t, []byte("value"), val)
			},
		},
		{
			name: "deploy and immediately drop database",
			fn: func(ctx context.Context, t *testing.T, registry *registry.Registry, recovery bool) {
				err := registry.Create(ctx, "db1")
				require.NoError(t, err)

				err = registry.Delete(ctx, "db1")
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		// testing regular registry operations
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			for _, dbid := range tt.initialDbs {
				pool := newMockPool(dbid)
				tracker.dbs[filepath.Join(dir, dbid)] = pool
			}

			registry, err := registry.NewRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			idempotencyKey := []byte("idempotencyKey")
			err = registry.Begin(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, false)

			_, err = registry.Commit(ctx, idempotencyKey)
			require.NoError(t, err)

			err = registry.Close()
			require.NoError(t, err)

			for _, dbid := range tt.hasPools {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.True(t, pool.closed)
			}

			for dbid, executions := range tt.hasExecutions {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.Equal(t, executions, pool.executed)
			}
		})

		// testing failure before Commit
		t.Run(tt.name+"-failure-before-commit", func(t *testing.T) {
			ctx := context.Background()
			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			for _, dbid := range tt.initialDbs {
				pool := newMockPool(dbid)
				tracker.dbs[filepath.Join(dir, dbid)] = pool
			}

			registry, err := registry.NewRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			idempotencyKey := []byte("idempotencyKey")
			err = registry.Begin(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, false)

			err = registry.Cancel(ctx)
			require.NoError(t, err)

			tracker.wipeUncommitted()
			tracker.wipePools()

			err = registry.Begin(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, false)

			_, err = registry.Commit(ctx, idempotencyKey)
			require.NoError(t, err)

			err = registry.Close()
			require.NoError(t, err)

			for _, dbid := range tt.hasPools {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.True(t, pool.closed)
			}

			for dbid, executions := range tt.hasExecutions {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.Equal(t, executions, pool.executed)
			}
		})

		// testing failure before Commit, and using recovery
		t.Run(tt.name+"-failure-before-commit-using-recovery", func(t *testing.T) {
			ctx := context.Background()
			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			for _, dbid := range tt.initialDbs {
				pool := newMockPool(dbid)
				tracker.dbs[filepath.Join(dir, dbid)] = pool
			}

			registry, err := registry.NewRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			idempotencyKey := []byte("idempotencyKey")
			err = registry.Begin(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, false)

			err = registry.Cancel(ctx)
			require.NoError(t, err)

			tracker.wipeUncommitted()
			tracker.wipePools()

			err = registry.BeginRecovery(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, true)

			_, err = registry.Commit(ctx, idempotencyKey)
			require.NoError(t, err)

			err = registry.Close()
			require.NoError(t, err)

			for _, dbid := range tt.hasPools {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.True(t, pool.closed)
			}

			for dbid, executions := range tt.hasExecutions {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.Equal(t, executions, pool.executed)
			}
		})

		// testing recovery on committed databases has no change
		// since the database is already committed, we should not see any changes
		t.Run(tt.name+"-recovery-on-committed-has-no-change", func(t *testing.T) {
			ctx := context.Background()
			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			for _, dbid := range tt.initialDbs {
				pool := newMockPool(dbid)
				tracker.dbs[filepath.Join(dir, dbid)] = pool
			}

			registry, err := registry.NewRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			idempotencyKey := []byte("idempotencyKey")
			err = registry.Begin(ctx, idempotencyKey)
			require.NoError(t, err)

			tt.fn(ctx, t, registry, false)

			_, err = registry.Commit(ctx, idempotencyKey)
			require.NoError(t, err)

			for i := 0; i < 10; i++ {
				err = registry.BeginRecovery(ctx, idempotencyKey)
				require.NoError(t, err)

				tt.fn(ctx, t, registry, true)

				_, err = registry.Commit(ctx, idempotencyKey)
				require.NoError(t, err)
			}

			err = registry.Close()
			require.NoError(t, err)

			for _, dbid := range tt.hasPools {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.True(t, pool.closed)
			}

			for dbid, executions := range tt.hasExecutions {
				pool, ok := tracker.getPool(dir, dbid)
				require.True(t, ok)
				require.Equal(t, executions, pool.executed)
			}
		})
	}
}

func Test_RegistryUncommitted(t *testing.T) {
	type testCase struct {
		name string
		fn   func(ctx context.Context, t *testing.T, registry *registry.Registry)
	}

	tests := []testCase{
		// {
		// 	name: "",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			registry, err := registry.NewRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			tt.fn(ctx, t, registry)
		})
	}
}

// poolTracker is a mock pool tracker.
type poolTracker struct {
	dbs map[string]*mockPool
}

// wipeUncommitted wipes the statements if the connection is not committed.
func (p *poolTracker) wipeUncommitted() {
	for _, pool := range p.dbs {
		pool.writer.wipeUncommitted()
	}
}

// wipePools wipes the pools of statements.
func (p *poolTracker) wipePools() {
	for _, pool := range p.dbs {
		pool.wipe()
	}
}

var _ registry.Filesystem = (*poolTracker)(nil)

func (p *poolTracker) getPool(path string, dbid string) (*mockPool, bool) {
	pool, ok := p.dbs[filepath.Join(path, dbid)]
	return pool, ok
}

func (p *poolTracker) Remove(path string) error {
	_, ok := p.dbs[path]
	if !ok {
		return fmt.Errorf("mockPoolTracker: database %q not found", path)
	}

	delete(p.dbs, path)
	return nil
}

func (p *poolTracker) Open(ctx context.Context, dbid string, create bool) (registry.Pool, error) {
	if create {
		if _, ok := p.dbs[dbid]; ok {
			return nil, fmt.Errorf("mockPoolTracker: database %q already exists", dbid)
		}

		pool := newMockPool(dbid)

		p.dbs[dbid] = pool

		return pool, nil
	}

	pool, ok := p.dbs[dbid]
	if !ok {
		return nil, fmt.Errorf("mockPoolTracker: database %q not found", dbid)
	}
	pool.closed = false

	return pool, nil
}

func (p *poolTracker) ForEachFile(path string, fn func(string) error) error {
	for fileName := range p.dbs {
		err := fn(filepath.Base(fileName))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *poolTracker) MkdirAll(path string, perms fs.FileMode) error {
	return nil
}

func (p *poolTracker) Rename(oldpath string, newpath string) error {
	file, ok := p.dbs[oldpath]
	if !ok {
		return fmt.Errorf("mockPoolTracker: database %q not found", oldpath)
	}

	delete(p.dbs, oldpath)

	p.dbs[newpath] = file

	return nil
}

type mockPool struct {
	onCommit   func()
	closed     bool
	writerOpen bool
	writer     *mockConnection
	executed   []executedStmt
}

func newMockPool(name string, onCommit ...func(string)) *mockPool {
	return &mockPool{
		onCommit: func() {
			for _, fn := range onCommit {
				fn(name)
			}
		},
		writer: &mockConnection{
			kv:       map[string][]byte{},
			executed: []executedStmt{},
		},
		executed: []executedStmt{},
	}
}

var _ registry.Pool = (*mockPool)(nil)

func (m *mockPool) Close() error {
	m.closed = true
	return nil
}

func (m *mockPool) BlockReaders(t time.Duration) func() {
	if m.closed {
		return nil
	}
	return func() {

	}
}

func (m *mockPool) Reader(p0 context.Context) (sql.ReturnableConnection, error) {
	if m.closed {
		return nil, fmt.Errorf("mockPool: already closed")
	}
	return &mockConnection{
		kv:       map[string][]byte{},
		executed: []executedStmt{},
	}, nil
}

func (m *mockPool) Writer() (sql.ReturnableConnection, error) {
	if m.closed {
		return nil, fmt.Errorf("mockPool: already closed")
	}

	if m.writerOpen {
		return nil, fmt.Errorf("mockPool: writer already open")
	}

	m.writerOpen = true
	m.writer.onClose = func() {
		m.writerOpen = false
	}
	m.writer.onCommit = m.onCommit

	return m.writer, nil
}

func (m *mockPool) CreateSession() (sql.Session, error) {
	return &mockSession{}, nil
}

func (m *mockPool) Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error) {
	m.executed = append(m.executed, executedStmt{
		stmt:   stmt,
		params: args,
	})
	return nil, nil
}

func (m *mockPool) Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error) {
	return nil, nil
}

func (m *mockPool) Savepoint() (sql.Savepoint, error) {
	return &mockSavepoint{}, nil
}

func (m *mockPool) Set(ctx context.Context, key []byte, value []byte) error {
	m.writer.kv[string(key)] = value
	return nil
}

func (m *mockPool) Get(ctx context.Context, key []byte, sync bool) ([]byte, error) {
	val, ok := m.writer.kv[string(key)]
	if !ok {
		return nil, nil
	}

	return val, nil
}

// wipeUncommitted wipes the statements
func (m *mockPool) wipe() {
	m.executed = []executedStmt{}
}

type mockConnection struct {
	onCommit      func()
	executed      []executedStmt
	kv            map[string][]byte
	onClose       func()
	committed     bool
	savepointOpen bool
}

type executedStmt struct {
	stmt   string
	params map[string]any
}

func (m *mockConnection) Execute(ctx context.Context, stmt string, params map[string]any) (sql.Result, error) {
	m.executed = append(m.executed, executedStmt{
		stmt:   stmt,
		params: params,
	})
	return &sql.EmptyResult{}, nil
}

func (m *mockConnection) Set(ctx context.Context, key []byte, value []byte) error {
	m.kv[string(key)] = value
	return nil
}

func (m *mockConnection) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, ok := m.kv[string(key)]
	if !ok {
		return nil, nil
	}

	return val, nil
}

func (m *mockConnection) Return() {
	if m.onClose != nil {
		m.onClose()
	}
}

func (m *mockConnection) Savepoint() (sql.Savepoint, error) {
	if m.savepointOpen {
		return nil, fmt.Errorf("mockConnection: savepoint already open")
	}

	return &mockSavepoint{
		commitFn: func() {
			m.onCommit()
			m.savepointOpen = false
			m.committed = true
		},
		rollbackFn: func() {
			m.savepointOpen = false
		},
	}, nil
}

// wipeUncommitted wipes the statements if the connection is not committed.
func (m *mockConnection) wipeUncommitted() {
	if !m.committed {
		m.executed = []executedStmt{}
	}
}

// CreateSession creates a session.
func (m *mockConnection) CreateSession() (sql.Session, error) {
	return &mockSession{}, nil
}

func (m *mockConnection) Close() error {
	return nil
}

func (m *mockConnection) Delete(ctx context.Context, key []byte) error {
	delete(m.kv, string(key))
	return nil
}

type mockSavepoint struct {
	commitFn   func()
	rollbackFn func()
}

func (m *mockSavepoint) Commit() error {
	if m.commitFn != nil {
		m.commitFn()
	}
	return nil
}

func (m *mockSavepoint) Rollback() error {
	if m.rollbackFn != nil {
		m.rollbackFn()
	}
	return nil
}

type mockSession struct{}

func (m *mockSession) Delete() error {
	return nil
}

func (m *mockSession) ChangesetID(ctx context.Context) ([]byte, error) {
	return []byte("changeset"), nil
}

/*
func Test_RegistryUncommitted(t *testing.T) {
	type testCase struct {
		name string
		fn   func(ctx context.Context, t *testing.T, registry *registry.Registry)
	}

	tests := []testCase{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tracker := &poolTracker{
				dbs: map[string]*mockPool{},
			}

			dir := "dir"

			registry, err := registry.OpenRegistry(ctx, tracker.Open, dir, registry.WithFilesystem(tracker), registry.WithReaderWaitTimeout(time.Duration(1)*time.Microsecond))
			require.NoError(t, err)

			idempotencyKey := []byte("idempotencyKey")

			tt.fn(ctx, t, registry)
		})
	}
}*/
