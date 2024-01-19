package pg

// This file defines a simple "replication monitor" for:
//  - listening for end-of-commit WAL data messages from a logical replication slot
//  - publishing updates with the message to a subscriber of a sequenced tx number
//
// It is designed for the DB type and is not intended to be used more generally.
// As such, none of this is exported.

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
)

// decodeCommitPayload extracts the seq value and commit hash from the data
// received from the logical replication message stream (see captureRepl).
func decodeCommitPayload(cid []byte) (int64, []byte, error) {
	if len(cid) <= 8 {
		return 0, nil, errors.New("invalid commit ID length")
	}
	seq := int64(binary.BigEndian.Uint64(cid))
	commitID := make([]byte, len(cid)-8)
	copy(commitID, cid[8:])
	return seq, commitID, nil
}

// replMon is the "replication monitor" that sits between the DB type and the
// receiver goroutine listening on a postgres replication slot. This is not
// exported for general use, but a consumer will use the recvID method, the
// errChan, and the done chan to interact.
type replMon struct {
	conn    *pgconn.PgConn
	errChan chan error
	quit    context.CancelFunc
	done    chan struct{}

	mtx      sync.Mutex
	results  map[int64][]byte
	promises map[int64]chan []byte
}

// newReplMon creates a new connection and logical replication data monitor, and
// immediately starts receiving messages from the host. A consumer should
// request a commit ID promise using the recvID method prior to committing a
// transaction.
func newReplMon(ctx context.Context, host, port, user, pass, dbName string, schemaFilter func(string) bool) (*replMon, error) {
	ctx, quit := context.WithCancel(ctx)
	conn, err := replConn(ctx, host, port, user, pass, dbName)
	if err != nil {
		quit()
		return nil, err
	}

	commitChan, errChan, err := startRepl(ctx, conn, "kwild_repl", "kwild_repl", schemaFilter) // todo: config publication name
	if err != nil {
		quit()
		conn.Close(ctx)
		return nil, err
	}

	rm := &replMon{
		conn:     conn,
		errChan:  errChan,
		quit:     quit,
		done:     make(chan struct{}),
		results:  make(map[int64][]byte),
		promises: make(map[int64]chan []byte),
	}

	go func() {
		defer quit()
		defer close(rm.done)
		defer conn.Close(ctx)

		for cid := range commitChan {
			// decode seq,chash
			seq, cHash, err := decodeCommitPayload(cid)
			if err != nil {
				rm.errChan <- fmt.Errorf("invalid commit payload encoding: %w", err)
				return
			}
			// if promise exists, send it, otherwise put it in the results map
			rm.mtx.Lock()
			if p, ok := rm.promises[seq]; ok {
				p <- cHash
				delete(rm.promises, seq)
			} else {
				rm.results[seq] = cHash
			}
			rm.mtx.Unlock()
		}
	}()

	return rm, nil
}

// testing this approach so that there can be multiple receivers, and the commit
// ID is guaranteed to pertain to the requested sequence number.
func (rm *replMon) recvID(seq int64) chan []byte {
	c := make(chan []byte, 1)

	// Ensure a commit ID can be promised before we give one.
	select {
	case <-rm.done:
		close(c)
		return c
	default:
	}

	// first check if the results is already in the map, otherwise make the
	// promise and store it
	rm.mtx.Lock()
	defer rm.mtx.Unlock()
	if cHash, ok := rm.results[seq]; ok {
		delete(rm.results, seq)
		c <- cHash
		return c
	}

	rm.promises[seq] = c // maybe panic if one already exists, indicating program logic error

	return c
}

func (rm *replMon) stop() {
	rm.quit()
	<-rm.done
	// rm.conn.Close(context.Background())
}
