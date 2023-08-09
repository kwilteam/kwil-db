package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"go.uber.org/zap"
)

type EngineSession struct {
	trackedDatasets map[string]*trackedDataset

	log log.Logger
}

// BeginCommit begins a session for each of the datasets that are being tracked by this engineSession.
func (e *EngineSession) BeginCommit(ctx context.Context) (err error) {
	defer func() {
		var err2 error
		if err != nil {
			err2 = e.untrackAll()
		}
		err = errors.Join(err, err2)
	}()

	for dbid, dataset := range e.trackedDatasets {
		session, err := dataset.db.CreateSession()
		if err != nil {
			return err
		}

		savepoint, err := dataset.db.Savepoint()
		if err != nil {
			return errors.Join(err, session.Delete())
		}

		e.trackedDatasets[dbid] = &trackedDataset{

			session:   session,
			savepoint: savepoint,
		}
	}

	return nil
}

// EndCommit commits the changes for each of the sessions that were created by this engineSession.
// It will rollback all of the savepoints and delete all of the sessions.
func (e *EngineSession) EndCommit(ctx context.Context, appender func([]byte) error) (commitId []byte, err error) {
	defer func() {
		err2 := e.deleteAndRollbackAll()
		if err2 != nil {
			e.log.Error("error rolling back savepoints", zap.Error(err2))
		}
	}()

	var idContent []byte

	for dbid, ds := range e.trackedDatasets {
		changeset, err := ds.session.GenerateChangeset()
		if err != nil {
			return nil, err
		}
		defer changeset.Close()

		id, err := changeset.ID()
		if err != nil {
			return nil, err
		}
		idContent = append(idContent, id...)

		data, err := changeset.Export()
		if err != nil {
			return nil, err
		}

		bts, err := serializeEngineChangeset(&engineChangeset{
			dbid:      dbid,
			changeset: data,
		})
		if err != nil {
			return nil, err
		}

		err = appender(bts)
		if err != nil {
			return nil, err
		}
	}

	return crypto.Sha256(idContent), nil
}

// BeginApply begins a session for each of the datasets that are being tracked by this engineSession.
func (e *EngineSession) BeginApply(ctx context.Context) error {
	for dbid, tracked := range e.trackedDatasets {
		if tracked.savepoint != nil {
			return fmt.Errorf("savepoint was open unexpectedly")
		}

		savepoint, err := tracked.db.Savepoint()
		if err != nil {
			return err
		}

		tracked.savepoint = savepoint

		e.trackedDatasets[dbid] = tracked
	}

	return nil
}

// Apply applies the changeset to the dataset.
func (e *EngineSession) Apply(ctx context.Context, changes []byte) error {
	changeset, err := deserializeEngineChangeset(changes)
	if err != nil {
		return err
	}

	tracked, ok := e.trackedDatasets[changeset.dbid]
	if !ok {
		return fmt.Errorf("cannot apply changeset for untracked dataset %v", changeset.dbid)
	}

	return tracked.db.ApplyChangeset(bytes.NewReader(changeset.changeset))
}

// EndApply commits the savepoints for each of the sessions that were created by this engineSession.
// It also checkpoints the sql wals.
func (e *EngineSession) EndApply(ctx context.Context) error {
	for _, tracked := range e.trackedDatasets {
		err := tracked.savepoint.Commit()
		if err != nil {
			return err
		}

		err = tracked.db.CheckpointWal()
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *EngineSession) Cancel(ctx context.Context) {
	err := e.untrackAll()
	if err != nil {
		e.log.Error("error untracking all while cancelling engine session", zap.Error(err))
	}
}

// untrackAll deletes all of the sessions and rollsback all savepoints that were created by this engineSession.
// it then clears the trackedDatasets map.
func (e *EngineSession) untrackAll() error {
	err := e.deleteAndRollbackAll()

	e.trackedDatasets = make(map[string]*trackedDataset)

	return err
}

// deleteAndRollbackAll deletes all of the sessions and rollsback all savepoints that were created by this engineSession.
func (e *EngineSession) deleteAndRollbackAll() error {
	errs := []error{}
	for _, ds := range e.trackedDatasets {
		err := ds.savepoint.Rollback()
		if err != nil {
			errs = append(errs, err)
		}

		err = ds.session.Delete()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// engineChangeset is a changeset that is tracked stored in the commit.
type engineChangeset struct {
	dbid      string
	changeset []byte
}

// serializeEngineChangeset serializes an engineChangeset into a byte slice.
func serializeEngineChangeset(changeset *engineChangeset) ([]byte, error) {
	var buf bytes.Buffer

	dbidLen := len(changeset.dbid)
	if dbidLen > 255 {
		return nil, errors.New("dbid too long to serialize")
	}

	// length of dbid
	buf.WriteByte(byte(dbidLen))

	// dbid
	buf.WriteString(changeset.dbid)

	// changeset
	buf.Write(changeset.changeset)

	return buf.Bytes(), nil
}

// deserializeEngineChangeset deserializes an engineChangeset from a byte slice.
func deserializeEngineChangeset(bts []byte) (*engineChangeset, error) {
	var changeset engineChangeset
	if len(bts) == 0 {
		return nil, errors.New("cannot deserialize empty byte slice")
	}

	// length of dbid
	dbidLen := int(bts[0])

	if len(bts) < dbidLen+1 {
		return nil, errors.New("cannot deserialize changeset: byte slice is too short")
	}

	bts = bts[1:]

	// dbid
	changeset.dbid = string(bts[:dbidLen])

	bts = bts[dbidLen:]

	// changeset
	changeset.changeset = bts

	return &changeset, nil
}

// trackedDataset is a dataset that is tracked by an engineSession.
type trackedDataset struct {
	db        SqlDB
	session   sql.Session
	savepoint sql.Savepoint
}

type SqlDB interface {
	ApplyChangeset(reader io.Reader) error
	CreateSession() (sql.Session, error)
	Savepoint() (sql.Savepoint, error)
	CheckpointWal() error
}
