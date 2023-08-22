package atomic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"

	sqlddlgenerator "github.com/kwilteam/kwil-db/pkg/engine/atomic/sql-ddl-generator"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql"
)

// AtomicEngine is a basic multi-file database engine in which all operations are atomic
// and deterministic
type AtomicEngine struct {
	// deployedDatasets is a set of all dbids that are currently deployed
	deployedDatasets map[string]struct{}

	// usedDatabases tracks the dbid and database connections that are currently in use
	usedDatabases map[string]Database

	// datasetOpener is the underlying engine that is used to open datasets
	datasetOpener DatabaseOpener

	// requestQueue is a queue of requests that are waiting to be processed
	// this guarantees that all requests are processed in order
	// linearizability is extremely important to make the engine deterministic
	requestQueue chan struct{}

	// changeTracker is the change tracker that is used to track changes
	changeTracker changeTracker

	// statementConverter makes a statement deterministic
	// if it cannot, it returns an error
	statementConverter StatementParser

	// encoder is used to encode and decode arbitrary structs deterministically
	encoder DeterministicEncoderDecoder

	// inSession tracks whether or not we are currently in a session
	inSession bool

	commitPhase commitPhase

	waiter Waiter
}

// getDatabase gets a database from the underlying engine
// TODO: add LRU cache for database connections.  we do not want to handle this in the opener,
// since we need the same database connection across an entire session to guarantee atomicity
// we need to figure out how we can use the same db connection so that we can use sessions and changesets
func (s *AtomicEngine) getDatabase(ctx context.Context, dbid string) (Database, error) {
	_, ok := s.deployedDatasets[dbid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	db, err := s.datasetOpener.OpenDatabase(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrOpeningDataset, dbid)
	}

	return db, nil
}

func (s *AtomicEngine) ApplyChanges(changes []byte) error {
	s.requestQueue <- struct{}{}
	defer func() { <-s.requestQueue }()

	panic("TODO")
}

func (a *AtomicEngine) CreateDataset(ctx context.Context, dbid string) error {
	a.requestQueue <- struct{}{}
	defer func() { <-a.requestQueue }()

	_, ok := a.deployedDatasets[dbid]
	if ok {
		return fmt.Errorf("%w: %s", ErrDatasetExists, dbid)
	}

	_, err := a.datasetOpener.OpenDatabase(ctx, dbid)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrOpeningDataset, dbid)
	}

	// TODO: add the db to an active dataset cache
	a.deployedDatasets[dbid] = struct{}{}

	a.changeTracker.TrackChange(&change{
		ID:   []byte("create-dbid-" + dbid),
		DBID: dbid,
		Type: ctCreateDataset,
	})

	return nil
}

func (a *AtomicEngine) DropDataset(ctx context.Context, dbid string) error {
	a.requestQueue <- struct{}{}
	defer func() { <-a.requestQueue }()

	_, ok := a.deployedDatasets[dbid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	if !a.inSession {
		err := a.datasetOpener.DeleteDatabase(ctx, dbid)
		if err != nil {
			return err
		}
	}

	delete(a.deployedDatasets, dbid)

	a.changeTracker.TrackChange(&change{
		ID:   []byte("drop-dbid-" + dbid),
		DBID: dbid,
		Type: ctDeleteDataset,
	})

	return nil
}

// CreateTable creates a table on the specified database
// If the table already exists, it returns an error
func (a *AtomicEngine) CreateTable(ctx context.Context, dbid string, table *types.Table) error {
	a.requestQueue <- struct{}{}
	defer func() { <-a.requestQueue }()

	db, err := a.getDatabase(ctx, dbid)
	if err != nil {
		return err
	}

	sp, err := db.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	err = createTable(ctx, db, table)
	if err != nil {
		return err
	}

	tableBytes, err := a.encoder.Encode(table)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(tableBytes)

	a.changeTracker.TrackChange(&change{
		ID:   hash[:],
		DBID: dbid,
		Type: ctCreateTable,
		Data: tableBytes,
	})

	return sp.Commit()
}

// Execute executes a statement on the underlying engine
// It should only be used for executing DML statements
func (a *AtomicEngine) Execute(ctx context.Context, dbid string, sqlStatement string, args map[string]any) ([]map[string]any, error) {
	a.requestQueue <- struct{}{}
	defer func() { <-a.requestQueue }()

	db, err := a.getDatabase(ctx, dbid)
	if err != nil {
		return nil, err
	}

	stmt, err := a.prepareStatement(db, sqlStatement)
	if err != nil {
		return nil, err
	}

	session, err := db.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Delete()

	sp, err := db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer sp.Rollback()

	res, err := stmt.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	changeset, err := session.GenerateChangeset()
	if err != nil {
		return nil, err
	}
	defer changeset.Close()

	id, err := changeset.ID()
	if err != nil {
		return nil, err
	}

	changesetBytes, err := changeset.Export()
	if err != nil {
		return nil, err
	}

	a.changeTracker.TrackChange(&change{
		ID:   id,
		DBID: dbid,
		Type: ctExecuteStatement,
		Data: changesetBytes,
	})

	err = sp.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// applyChange applies a change to the underlying engine
// it should be idempotent
func (a *AtomicEngine) applyChange(ctx context.Context, change *change) error {
	switch change.Type {
	case ctCreateDataset:
		_, err := a.datasetOpener.OpenDatabase(ctx, change.DBID)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrOpeningDataset, change.DBID)
		}

		a.deployedDatasets[change.DBID] = struct{}{}
		return nil
	case ctDeleteDataset:
		delete(a.deployedDatasets, change.DBID)
		return a.datasetOpener.DeleteDatabase(ctx, change.DBID)
	case ctCreateTable:
		db, err := a.getDatabase(ctx, change.DBID)
		if err != nil {
			return err
		}

		table := &types.Table{}
		err = a.encoder.Decode(change.Data, table)
		if err != nil {
			return err
		}

		// create a savepoint since createTable is not atomic
		sp, err := db.Savepoint()
		if err != nil {
			return err
		}
		defer sp.Rollback()

		// attempt to create the table
		// if it already exists, we do not need to do anything
		err = createTable(ctx, db, table)
		if err != nil && err != ErrTableExists {
			return err
		}

		return sp.Commit()
	case ctExecuteStatement:
		db, err := a.getDatabase(ctx, change.DBID)
		if err != nil {
			return err
		}

		sp, err := db.Savepoint()
		if err != nil {
			return err
		}
		defer sp.Rollback()

		err = db.ApplyChangeset(bytes.NewReader(change.Data))
		if err != nil {
			return err
		}

		return sp.Commit()
	default:
		panic("unknown change type: " + string(change.Type))
	}
}

// TODO: add LRU cache for statement preparation
// this should be done by taking the hash of the unparsed statement, and using that as the key
// to the prepared statement (which is prepared based on the parsed statement)
func (a *AtomicEngine) prepareStatement(db Database, statement string) (sql.Statement, error) {
	converted, err := a.statementConverter(statement)
	if err != nil {
		return nil, err
	}

	return db.Prepare(converted)
}

func (a *AtomicEngine) BeginCommit(ctx context.Context) error {
	a.requestQueue <- struct{}{}
	defer func() { <-a.requestQueue }()
	panic("TODO")
}

func (a *AtomicEngine) EndCommit(ctx context.Context, appender func([]byte) error) error {

	panic("TODO")
}

// BeginApply begins an apply session
// It will block the request queue until EndApply or Cancel is called
func (a *AtomicEngine) BeginApply(ctx context.Context) error {
	done, err := a.waiter.Wait(ctx)
	if err != nil {
		return err
	}
	defer done()
}

type Waiter interface {
	// Wait waits until it is the callers turn to apply changes
	// It is returned an error if the context times out or if the buffer is full
	// it returns a function to signal that it is done
	Wait(ctx context.Context) (func(), error)
}

// Apply applies a set of changes to the underlying engine
// it does not block, and therefore should only be called after BeginApply
func (a *AtomicEngine) Apply(ctx context.Context, changes []byte) error {
	panic("TODO")
}

func (a *AtomicEngine) EndApply(ctx context.Context) error {
	if a.commitPhase != commitPhaseApply {
		return fmt.Errorf("%w: %s", ErrInvalidCommitPhase, a.commitPhase)
	}

}

func (a *AtomicEngine) Cancel(ctx context.Context) {
	if a.commitPhase == commitPhaseApply {
		defer func() { <-a.requestQueue }()
	}
}

func (a *AtomicEngine) ID(ctx context.Context) ([]byte, error) {
	panic("TODO")
}

type commitPhase uint8

const (
	commitPhaseNone commitPhase = iota
	commitPhaseCommit
	commitPhaseApply
)

// createTable creates a table on the provided database
// if it already exists, it returns an ErrTableExists error
func createTable(ctx context.Context, db Database, table *types.Table) error {
	exists, err := db.TableExists(ctx, table.Name)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("%w: %s", ErrTableExists, table.Name)
	}

	ddlStatements, err := sqlddlgenerator.GenerateDDL(table)
	if err != nil {
		return err
	}

	for _, ddlStatement := range ddlStatements {
		err = db.Execute(ctx, ddlStatement, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
