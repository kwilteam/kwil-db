package datasets

//Manages Block Sessions for all the sql databases
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/sql"
	gowal "github.com/tidwall/wal"
)

type SessionMgr struct {
	shadowWal      *gowal.Log
	accountSession *AccountSession
	dbSession      map[string]*DbSession
	prevAppHash    []byte
	// governanceSession *DbSession
}

type DbSession struct {
	session   sql.Session
	savepoint sql.Savepoint
	changeset []byte
}

type AccountSession struct {
	session   balances.Session
	savepoint balances.Savepoint
	changeset []byte
}

func NewSessionMgr(appHash []byte) (*SessionMgr, error) {
	CometHomeDir := os.Getenv("COMET_BFT_HOME")
	ShadowDBWalPath := filepath.Join(CometHomeDir, "data", "ShadowDb.wal")
	wal, err := gowal.Open(ShadowDBWalPath, nil)
	if err != nil {
		return nil, err
	}
	// This is done to reset the indexes to 1, [need this to avoid a bug with the truncation of tidwal]
	wal.Write(1, []byte("no-op"))
	wal.TruncateBack(1)

	return &SessionMgr{
		shadowWal:   wal,
		prevAppHash: appHash,
		dbSession:   make(map[string]*DbSession),
	}, nil
}

func (u *DatasetUseCase) InitalizeAppHash(appHash []byte) {
	u.sessionMgr.prevAppHash = appHash
}

func (u *DatasetUseCase) StartBlockSession() error {
	// Open Account store session
	err := u.accountSessionStart()
	if err != nil {
		return err
	}
	// TODO: Governance session

	// Open Datastore sessions
	dbids, err := u.engine.GetAllDatasets()
	if err != nil {
		return err
	}
	for _, dbid := range dbids {
		err = u.createDbSession(dbid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *DatasetUseCase) EndBlockSession() ([]byte, error) {
	// Retrieve and store the changesets in shadowWal
	u.persistChangesets()

	// Generate App Hash
	appHash := u.sessionMgr.generateAppHash()

	// Apply the changesets to the sqlDBs
	err := u.applyChangesets()
	if err != nil {
		return nil, err
	}
	u.sessionMgr.prevAppHash = appHash
	u.sessionMgr.accountSession = nil
	u.sessionMgr.dbSession = make(map[string]*DbSession)
	return appHash, nil
}

func (u *DatasetUseCase) accountSessionStart() error {
	// Open Account store session
	savepoint, err := u.accountStore.Savepoint()
	if err != nil {
		return err
	}

	session, err := u.accountStore.CreateSession()
	if err != nil {
		return err
	}

	u.sessionMgr.accountSession = &AccountSession{
		session:   session,
		savepoint: savepoint,
		changeset: nil,
	}
	return nil
}

func (u *DatasetUseCase) createDbSession(dbid string) error {
	ctx := context.Background()
	ds, err := u.engine.GetDataset(ctx, dbid)
	if err != nil {
		return err
	}

	savepoint, err := ds.Savepoint()
	if err != nil {
		return err
	}

	session, err := ds.CreateSession()
	if err != nil {
		return err
	}

	u.sessionMgr.dbSession[dbid] = &DbSession{
		session:   session,
		savepoint: savepoint,
		changeset: nil,
	}
	return nil
}

func (u *DatasetUseCase) removeDbSession(dbid string) error {
	delete(u.sessionMgr.dbSession, dbid)
	return nil
}

func (u *DatasetUseCase) persistChangesets() error {
	sessMgr := u.sessionMgr
	lastIdx, _ := sessMgr.shadowWal.LastIndex()
	idx := lastIdx + 1
	// Account store changesets
	err := u.endAccountStoreSession()
	if err != nil {
		return err
	}
	idx, err = sessMgr.writeChangeset("accounts", sessMgr.accountSession.changeset, idx)
	if err != nil {
		return err
	}

	// TODO: Governance store changesets

	// Datastore changesets
	for dbid := range u.sessionMgr.dbSession {
		err = u.endDbSession(dbid)
		if err != nil {
			return err
		}
		idx, err = sessMgr.writeChangeset(dbid, sessMgr.dbSession[dbid].changeset, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *DatasetUseCase) endAccountStoreSession() error {
	acctSession := u.sessionMgr.accountSession
	cs, err := acctSession.session.GenerateChangeset()
	if err != nil {
		return err
	}
	acctSession.changeset = cs
	err = acctSession.savepoint.Rollback()
	if err != nil {
		return err
	}
	acctSession.session.Delete()
	return nil
}

func (u *DatasetUseCase) endDbSession(dbid string) error {
	dbSession := u.sessionMgr.dbSession[dbid]
	cs, err := dbSession.session.GenerateChangeset()
	if err != nil {
		return err
	}
	dbSession.changeset = cs
	err = dbSession.savepoint.Rollback()
	if err != nil {
		return err
	}
	dbSession.session.Delete()
	return nil
}

func (u *DatasetUseCase) applyChangesets() error {
	wal := u.sessionMgr.shadowWal
	firstIdx, _ := wal.FirstIndex()
	lastIdx, _ := wal.LastIndex()
	idx := firstIdx + 1

	for idx <= lastIdx-1 {
		name, err := wal.Read(idx)
		if err != nil {
			return err
		}
		cs, err := wal.Read(idx + 1)
		if err != nil {
			return err
		}

		switch string(name) {
		case "accounts":
			fmt.Println("Account store changeset: ", string(cs))
			err = u.accountStore.ApplyChangeset(strings.NewReader(string(cs)))
			if err != nil {
				fmt.Println("Error in applying acctstore changeset: ", err)
				return err
			}
		case "governance":
			// TODO: For node join related changes
		default:
			ds, err := u.engine.GetDataset(context.Background(), string(name))
			if err != nil {
				return err
			}
			fmt.Println("Datastore changeset: ", string(name), "   ", string(cs))
			err = ds.ApplyChangeset(strings.NewReader(string(cs)))
			if err != nil {
				fmt.Println("Error in applying changeset: ", err)
				return err
			}
		}
		idx += 2
	}
	wal.TruncateBack(firstIdx)
	return nil
}

func (s *SessionMgr) writeChangeset(dbid string, cs []byte, idx uint64) (uint64, error) {
	if cs == nil {
		return idx, nil
	}
	batch := new(gowal.Batch)
	batch.Write(idx, []byte(dbid))
	batch.Write(idx+1, cs)
	return idx + 2, s.shadowWal.WriteBatch(batch)
}

func (s *SessionMgr) generateAppHash() []byte {
	/*
		Cumulative changeset for all the sql databases
		cumulativeCS = (AccountstoreCS + GovernanceStoreCS + DatastoreCS)
		changeSetHash = sha256(cumulativeCS)
		AppHash = sha256(prevAppHash + changeSetHash)
	*/
	cumulativeCS := ""

	// Account store changesets
	if s.accountSession.changeset != nil {
		cumulativeCS += string(s.accountSession.changeset)
	}

	// TODO: Governance store changesets

	// Datastore changesets (Sort the datastores by dbid and append the changesets in that order)
	dbids := make([]string, 0, len(s.dbSession))
	for dbid, session := range s.dbSession {
		if session.changeset == nil {
			continue
		}
		dbids = append(dbids, dbid)
	}

	if len(dbids) != 0 {
		sort.Strings(dbids)
		for _, dbid := range dbids {
			cumulativeCS += string(s.dbSession[dbid].changeset)
		}
	}

	if cumulativeCS == "" {
		return s.prevAppHash
	}

	dbHash := crypto.Sha256([]byte(cumulativeCS))
	appHash := crypto.Sha256(append(s.prevAppHash, dbHash...))

	return appHash[:]
}
