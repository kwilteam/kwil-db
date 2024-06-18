//go:build pglive

package pg

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/utils/random"
)

func Test_repl(t *testing.T) {
	UseLogger(log.NewStdOut(log.DebugLevel))
	host, port, user, pass, dbName := "127.0.0.1", "5432", "kwild", "kwild", "kwil_test_db"

	ctx := context.Background()
	conn, err := replConn(ctx, host, port, user, pass, dbName)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	sysident, err := pglogrepl.IdentifySystem(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("SystemID:", sysident.SystemID, "Timeline:", sysident.Timeline,
		"XLogPos:", sysident.XLogPos, "DBName:", sysident.DBName)

	deadline, exists := t.Deadline()
	if !exists {
		deadline = time.Now().Add(2 * time.Minute)
	}

	ctx, cancel := context.WithDeadline(ctx, deadline.Add(-time.Second*5))
	defer cancel()
	connQ, err := pgx.Connect(ctx, connString(host, port, user, pass, dbName, false))
	if err != nil {
		t.Fatal(err)
	}
	_, err = connQ.Exec(ctx, sqlUpdateSentrySeq, 0)
	if err != nil {
		t.Fatal(err)
	}

	schemaFilter := func(string) bool { return true } // capture changes from all namespaces

	const publicationName = "kwild_repl"
	var slotName = publicationName + random.String(8)
	commitChan, errChan, quit, err := startRepl(ctx, conn, publicationName, slotName, schemaFilter, &changesetIoWriter{})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("replication slot started and listening")

	_, err = connQ.Exec(ctx, `DROP TABLE IF EXISTS blah`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = connQ.Exec(ctx, `CREATE TABLE IF NOT EXISTS blah (id BYTEA PRIMARY KEY, stuff TEXT NOT NULL, val INT8)`)
	if err != nil {
		t.Fatal(err)
	}

	wantCommitHash, _ := hex.DecodeString("cb390afbf808256307ee0927999805ee3d5af193772e2c9b71823fbc1fe8867f")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case cid := <-commitChan:
				_, commitHash, err := decodeCommitPayload(cid)
				if err != nil {
					t.Errorf("invalid commit payload encoding: %v", err)
					return
				}
				// t.Logf("Commit HASH: %x\n", commitHash)
				if !bytes.Equal(commitHash, wantCommitHash) {
					t.Errorf("commit hash mismatch, got %x, wanted %x", commitHash, wantCommitHash)
				}
				quit()
			case err := <-errChan:
				if errors.Is(err, context.Canceled) {
					return
				}
				if errors.Is(err, context.DeadlineExceeded) {
					t.Error("timeout")
					return
				}
				if err != nil {
					t.Error(err)
					quit()
				}
				return
			}
		}
	}()

	tx, err := connQ.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Exec(ctx, `insert INTO blah values ( '{11}', 'woot' , 42);`)
	tx.Exec(ctx, `update blah SET stuff = 6, id = '{13}', val=41 where id = '{10}';`)
	tx.Exec(ctx, `update blah SET stuff = 33;`)
	tx.Exec(ctx, `delete FROM blah where id = '{11}';`)
	// sends on commitChan are only expected from sequenced transactions.
	// Bump seq in the sentry table!
	_, err = tx.Exec(ctx, sqlUpdateSentrySeq, 1)
	if err != nil {
		t.Fatal(err)
	}

	err = tx.Commit(ctx) // this triggers the send
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	connQ.Close(ctx)
}
