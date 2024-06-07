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

	"github.com/kwilteam/kwil-db/core/utils/random"
)

// publicationName is the name of the publication required for logical
// replication.
const publicationName = "kwild_repl"

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
	conn *pgconn.PgConn
	quit context.CancelFunc
	done chan struct{} // termination broadcast channel
	err  error         // specific error, safe to read after done is closed

	mtx      sync.Mutex
	results  map[int64][]byte // results should generally be unused as pg.DB will request a promise before commit
	promises map[int64]chan []byte
}

// newReplMon creates a new connection and logical replication data monitor, and
// immediately starts receiving messages from the host. A consumer should
// request a commit ID promise using the recvID method prior to committing a
// transaction.
func newReplMon(ctx context.Context, host, port, user, pass, dbName string, schemaFilter func(string) bool) (*replMon, error) {
	conn, err := replConn(ctx, host, port, user, pass, dbName)
	if err != nil {
		return nil, err
	}

	var slotName = publicationName + random.String(8) // arbitrary, so just avoid collisions
	commitChan, errChan, quit, err := startRepl(ctx, conn, publicationName, slotName, schemaFilter)
	if err != nil {
		quit()
		conn.Close(ctx)
		return nil, err
	}

	rm := &replMon{
		conn:     conn,
		quit:     quit,
		done:     make(chan struct{}),
		results:  make(map[int64][]byte),
		promises: make(map[int64]chan []byte),
	}

	go func() {
		defer close(rm.done)
		defer quit()
		defer conn.Close(context.Background())

		for cid := range commitChan { // until commitChan is closed
			// decode seq,chash
			seq, cHash, err := decodeCommitPayload(cid)
			if err != nil {
				rm.err = fmt.Errorf("invalid commit payload encoding: %w", err)
				return // quit() will terminate startRepl
			}
			// if promise exists, send it, otherwise put it in the results map
			rm.mtx.Lock()
			if p, ok := rm.promises[seq]; ok {
				p <- cHash
				delete(rm.promises, seq)
			} else {
				// This is unexpected since pg.DB will call recvID first. If we are
				// in this `else`, it is to be discarded, from another connection.
				logger.Warnf("Received commit ID for seq %d BEFORE recvID", seq)
				rm.results[seq] = cHash
			}
			rm.mtx.Unlock()
		}

		// commitChan was closed by the replication stream goroutine, so we
		// expect a cause from its errChan. It could just be context.Canceled
		// from a clean shutdown, or it could be something pathological.
		rm.err = <-errChan // send guaranteed before commitChan closed
	}()

	return rm, nil
}

// this channel-based approach is so that the commit ID is guaranteed to pertain
// to the requested sequence number.
func (rm *replMon) recvID(seq int64) (chan []byte, bool) {
	// Ensure a commit ID can be promised before we give one.
	select {
	case <-rm.done:
		return nil, false
	default:
	}

	c := make(chan []byte, 1)

	// first check if the results is already in the map, otherwise make the
	// promise and store it
	rm.mtx.Lock()
	defer rm.mtx.Unlock()
	if cHash, ok := rm.results[seq]; ok {
		// The intended use is to do recvID BEFORE
		logger.Warnf("recvID with EXISTING result for sequence %d", seq)
		delete(rm.results, seq)
		c <- cHash
		return c, true
	}

	if _, have := rm.promises[seq]; have {
		logger.Errorf("Commit ID promise for sequence %d ALREADY EXISTS", seq)
	}
	rm.promises[seq] = c // maybe panic if one already exists, indicating program logic error

	return c, true
}

func (rm *replMon) stop() {
	rm.quit()
	<-rm.done
	// rm.conn.Close(context.Background())
}
