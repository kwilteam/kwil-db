package atomic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/atomic/queue"
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
	requestQueue *queue.Queue

	// changeTracker is the change tracker that is used to track changes
	changeTracker changeTracker

	// statementConverter makes a statement deterministic
	// if it cannot, it returns an error
	statementConverter StatementParser

	// encoder is used to encode and decode arbitrary structs deterministically
	encoder DeterministicEncoderDecoder

	// session holds information for the current session
	session *sessionInfo

	commitPhase commitPhase

	// applyFinisher is a function that should be called at the end
	// of the apply phase.  it should either be called in Cancel or EndApply
	applyFinisher func()
}

// getDatabase gets a database connection for the provided dbid.
// if the database is already used in the session, it will get the connection from the session cache.
func (s *AtomicEngine) openDatabase(ctx context.Context, dbid string) (Database, error) {
	_, ok := s.deployedDatasets[dbid]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	if s.commitPhase.inSession() {
		db, ok := s.usedDatabases[dbid]
		if ok {
			return db, nil
		}
	}

	db, err := s.datasetOpener.OpenDatabase(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrOpeningDataset, dbid)
	}

	if s.commitPhase.inSession() {
		err = s.session.RegisterDatabase(ctx, dbid, db)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// CreateDataset creates a dataset on the underlying engine
// It also registers the dataset with the session
func (a *AtomicEngine) CreateDataset(ctx context.Context, dbid string) error {
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return err
	}
	defer done()

	_, ok := a.deployedDatasets[dbid]
	if ok {
		return fmt.Errorf("%w: %s", ErrDatasetExists, dbid)
	}

	db, err := a.datasetOpener.OpenDatabase(ctx, dbid)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrOpeningDataset, dbid)
	}

	err = a.session.RegisterDatabase(ctx, dbid, db)
	if err != nil {
		return err
	}

	a.deployedDatasets[dbid] = struct{}{}

	a.changeTracker.TrackChange(&change{
		ID:   []byte("create-dbid-" + dbid),
		DBID: dbid,
		Type: ctCreateDataset,
	})

	return nil
}

func (a *AtomicEngine) DropDataset(ctx context.Context, dbid string) error {
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return err
	}
	defer done()

	_, ok := a.deployedDatasets[dbid]
	if !ok {
		return fmt.Errorf("%w: %s", ErrDatasetNotFound, dbid)
	}

	// if not in session, delete immediately
	if !a.commitPhase.inSession() {
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
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return err
	}
	defer done()

	db, err := a.openDatabase(ctx, dbid)
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
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	db, err := a.openDatabase(ctx, dbid)
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
		db, err := a.openDatabase(ctx, change.DBID)
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
		db, err := a.openDatabase(ctx, change.DBID)
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
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return err
	}
	defer done()

	panic("TODO")
}

func (a *AtomicEngine) EndCommit(ctx context.Context, appender func([]byte) error) error {

	panic("TODO")
}

// BeginApply begins an apply session
// It will block the request queue until EndApply or Cancel is called
func (a *AtomicEngine) BeginApply(ctx context.Context) error {
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return err
	}
	a.applyFinisher = done

	panic("TODO")
}

// Apply applies a set of changes to the underlying engine
// it does not block, and therefore should only be called after BeginApply
func (a *AtomicEngine) Apply(ctx context.Context, changes []byte) error {

	panic("TODO")
}

// cleanupApply cleans up the apply phase finisher, if necessary
func (a *AtomicEngine) cleanupApply() {
	if a.applyFinisher != nil {
		// preventing race condition where we call the applyFinisher
		// and it gets called again before we set it to nil
		fn := a.applyFinisher
		a.applyFinisher = nil
		fn()
	}
}

// if EndApply is successful, it will call the applyFinisher
// if not, it will not call the applyFinisher, as it will be called in Cancel
func (a *AtomicEngine) EndApply(ctx context.Context) error {
	if a.commitPhase != commitPhaseApply {
		return fmt.Errorf("%w: current phase: %s.  expected: ", ErrInvalidCommitPhase, a.commitPhase, commitPhaseApply)
	}

	// TODO: implement

	// at the end:
	a.cleanupApply()
	return nil
}

func (a *AtomicEngine) Cancel(ctx context.Context) {
	// if we are in the apply phase, we need to call the applyFinisher
	if a.commitPhase == commitPhaseApply {
		a.cleanupApply()
	}

	// TODO: implement
}

func (a *AtomicEngine) ID(ctx context.Context) ([]byte, error) {
	done, err := a.requestQueue.Wait(ctx)
	if err != nil {
		return nil, err
	}
	defer done()

	panic("TODO")
}

type commitPhase uint8

const (
	commitPhaseNone commitPhase = iota
	commitPhaseCommit
	commitPhaseApply
)

// inSession returns true if the engine is currently in a session
func (c commitPhase) inSession() bool {
	return c != commitPhaseNone
}

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
