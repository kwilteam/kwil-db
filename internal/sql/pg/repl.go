package pg

// This file contains the low level functions for streaming and decoding WAL
// data messages from a logical replication slot, and digesting the messages
// pertaining to data updates (UPDATE, INSERT, DELETE, TRUNCATE) on a subset of
// namespaces (postgres schema). This is only used by the replMon in replmon.go
// via the DB type's outermost transaction handling. As such, none of this is
// exported or well generalized.
//
// It recognizes UPDATEs to a special kwild_internal.sentry table, and captures
// a sequence value to identify the committed transaction. If none was set, as
// would be done by the DB type, it remains -1.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kwilteam/kwil-db/core/log"
)

// replConn creates a new connection to a postgres host with the
// `replication=database` setting in the connection string so that it can be
// used to receive logical replication messages with WAL update data. This is
// low level, used by DB via newReplMon to prepare the connection for startRepl.
func replConn(ctx context.Context, host, port, user, pass, dbName string) (*pgconn.PgConn, error) {
	const repl = true
	connStr := connString(host, port, user, pass, dbName, repl)

	return pgconn.Connect(ctx, connStr)
}

// startRepl creates a replication slot and begins receiving data. Cancelling
// the context only cancels creation of the connection. Use the quit function to
// terminate the monitoring goroutine.
func startRepl(ctx context.Context, conn *pgconn.PgConn, publicationName, slotName string,
	schemaFilter func(string) bool, writer *changesetIoWriter) (chan []byte, chan error, context.CancelFunc, error) {
	// Create the replication slot and start postgres sending WAL data.
	startLSN, err := createRepl(ctx, conn, publicationName, slotName)
	if err != nil {
		return nil, nil, nil, err
	}

	// Launch the receiver goroutine, which will send commit digests and an
	// error on return.
	done := make(chan error, 1)

	// WARNING: there must be a commitHash receiver for every send. This is
	// coordinated by only sending commit IDs on this channel for transactions
	// containing a sequence number update on the internal sentry table. This
	// means: (1) there must only be one pg.DB instance per postgres database,
	// and (2) other unsequenced writers such as a pg.Pool must not make updates
	// to the sentry table that would cause a send with no receiver.
	commitHash := make(chan []byte, 1)

	// Tie captureRepl goroutine to a new context now that connections are established.
	ctx2, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(commitHash)
		done <- captureRepl(ctx2, conn, uint64(startLSN), commitHash, schemaFilter, writer)
	}()

	return commitHash, done, cancel, nil
}

func createRepl(ctx context.Context, conn *pgconn.PgConn, publicationName, slotName string) (pglogrepl.LSN, error) {
	sysident, err := pglogrepl.IdentifySystem(ctx, conn)
	if err != nil {
		return 0, err
	}

	logger.Debug("postgres IDENTIFY_SYSTEM", log.String("SystemID", sysident.SystemID),
		log.Int("Timeline", sysident.Timeline), log.String("XLogPos", sysident.XLogPos.String()),
		log.String("DBName", sysident.DBName))

	// const publicationName = "kwild_repl"
	// Creating the publication should be done with psql as a superuser when
	// creating the kwild database and role.
	//  e.g.
	//  CREATE USER kwild WITH SUPERUSER REPLICATION; -- verify: SELECT rolname, rolreplication FROM pg_roles WHERE rolreplication = true;
	//  CREATE DATABASE kwild OWNER kwild;
	//  -- then '\c kwild' to connect to the kwild database
	//  CREATE PUBLICATION kwild_repl FOR ALL TABLES; -- applies to connected DB! also, this can be auto if kwild user is superuser

	// slotRes, err := pglogrepl.CreateReplicationSlot(ctx, conn, slotName, "pgoutput",
	// 	pglogrepl.CreateReplicationSlotOptions{
	// 		Mode:      pglogrepl.LogicalReplication,
	// 		Temporary: true,
	// 	})

	// We do this manually with Exec so we can enable two-phase commit mode with
	// prepared transactions.
	sqlStartRepl := fmt.Sprintf("CREATE_REPLICATION_SLOT %s TEMPORARY LOGICAL pgoutput TWO_PHASE", slotName)
	slotRes, err := pglogrepl.ParseCreateReplicationSlot(conn.Exec(ctx, sqlStartRepl))
	if err != nil {
		return 0, err
	}
	slotLSN, _ := pglogrepl.ParseLSN(slotRes.ConsistentPoint)
	logger.Infof("Created logical replication slot %v at LSN %v (%d)",
		slotRes.SlotName, slotRes.ConsistentPoint, slotLSN)

	pluginArgs := []string{
		"proto_version '3'",
		"publication_names '" + publicationName + "'",
		"messages 'true'",
		"streaming 'true'",
	}
	err = pglogrepl.StartReplication(ctx, conn, slotName, sysident.XLogPos,
		pglogrepl.StartReplicationOptions{
			PluginArgs: pluginArgs,
			Mode:       pglogrepl.LogicalReplication,
		})
	if err != nil {
		return 0, fmt.Errorf("StartReplication failed: %w", err)
	}

	return sysident.XLogPos, nil
}

var zeroHash, _ = hex.DecodeString("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

// captureRepl begins receiving and decoding messages. Consider the conn to be
// hijacked after calling captureRepl, and do not use it or the stream can be
// broken. It decodeFullWal is true, it will return the entire wal serialized,
// instead of just the commit hash.
func captureRepl(ctx context.Context, conn *pgconn.PgConn, startLSN uint64, commitHash chan []byte,
	schemaFilter func(string) bool, writer *changesetIoWriter) error {
	if cap(commitHash) == 0 {
		return errors.New("buffered commit hash channel required")
	}

	clientXLogPos := pglogrepl.LSN(startLSN)
	standbyMessageTimeout := time.Second * 10
	nextStandbyMessageDeadline := time.Now().Add(standbyMessageTimeout)
	hasher := sha256.New()
	relations := map[uint32]*pglogrepl.RelationMessageV2{}

	var inStream bool
	var seq int64 = -1

	stats := new(walStats)

	// The following loop receives messages from the replication receiver
	// connection. This includes ALL message types, not just replication
	// messages. The message types are:
	//
	//  - CopyData includes data that may pertain to logical replication
	//    or just keepalive messages. This is the expected type.
	//  - CommandComplete is the signal from postgres that it is terminating
	//    the connection, such as if it is shutting down.
	//  - ErrorResponse is a fatal error that may contain any of the common
	//    "SQLSTATE" codes.
	//
	// Any other message type is logged, but the loop is otherwise
	// uninterrupted. A loop is required since there may be no concurrent use of
	// this low level connection.

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if time.Now().After(nextStandbyMessageDeadline) {
			err := pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: clientXLogPos})
			if err != nil {
				return fmt.Errorf("SendStandbyStatusUpdate failed: %w", err)
			}
			logger.Debugf("Sent Standby status message at %s (%d)\n", clientXLogPos, uint64(clientXLogPos))
			nextStandbyMessageDeadline = time.Now().Add(standbyMessageTimeout)
		}

		// Timeout ReceiveMessage to send the next standby status message.
		ctxStandby, cancel := context.WithDeadline(ctx, nextStandbyMessageDeadline)
		rawMsg, err := conn.ReceiveMessage(ctxStandby)
		cancel()
		if err != nil {
			if pgconn.Timeout(err) || errors.Is(err, context.DeadlineExceeded) {
				continue // nextStandbyMessageDeadline hit, time to send next standby status message
			}
			return fmt.Errorf("ReceiveMessage failed: %w", err)
		}

		var msg *pgproto3.CopyData

		switch msgT := rawMsg.(type) {
		case *pgproto3.CopyData:
			msg = msgT
		case *pgproto3.CommandComplete:
			return fmt.Errorf("postgresql has been prematurely stopped")
		case *pgproto3.ErrorResponse:
			return fmt.Errorf("received Postgres WAL stream error: %+v", msgT)
		default:
			logger.Warnf("Received unexpected message: %T\n", rawMsg)
			continue
		}

		switch msg.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			pkm, err := pglogrepl.ParsePrimaryKeepaliveMessage(msg.Data[1:])
			if err != nil {
				return fmt.Errorf("ParsePrimaryKeepaliveMessage failed: %w", err)
			}
			logger.Debug("primary keepalive msg", log.String("ServerWALEnd", pkm.ServerWALEnd.String()),
				log.String("ServerTime:", pkm.ServerTime.String()), log.Bool("ReplyRequested:", pkm.ReplyRequested))
			if pkm.ServerWALEnd > clientXLogPos {
				clientXLogPos = pkm.ServerWALEnd
			}
			if pkm.ReplyRequested {
				nextStandbyMessageDeadline = time.Time{}
			}

		case pglogrepl.XLogDataByteID:
			xld, err := pglogrepl.ParseXLogData(msg.Data[1:])
			if err != nil {
				return fmt.Errorf("ParseXLogData failed: %w", err)
			}

			// TODO: delete this check. This is a sanity check for some new logic
			// if len(writer.data) != 0 {
			// 	panic("DELETEME: changesetIoWriter not empty")
			// }

			commit, anySeq, err := decodeWALData(hasher, xld.WALData, relations, &inStream, stats, schemaFilter, writer)
			if err != nil {
				return fmt.Errorf("decodeWALData failed: %w", err)
			}
			if anySeq != -1 { // the seq update at the beginning of a transaction
				if seq != -1 {
					return fmt.Errorf("sequence already set")
				}
				seq = anySeq // the magic sentry table UPDATE that precedes commit
			}

			var lsnDelta uint64
			if xld.WALStart > clientXLogPos {
				lsnDelta = uint64(xld.WALStart - clientXLogPos)
				clientXLogPos = xld.WALStart
			}

			// logger.Debugf("XLogData (in stream? %v) => WALStart %s ServerWALEnd %s\n",
			// 	inStream, xld.WALStart, xld.ServerWALEnd)

			if commit {
				cHash := hasher.Sum(nil)
				hasher.Reset() // hasher = sha256.New()

				// Only send the commit ID on the commitHash channel if this was
				// a tracked commit, which includes a sequence number update on
				// the internal sentry table that indicates it was created by
				// the pg.DB type.
				if seq == -1 {
					logger.Debugf("Commit hash %x (unsequenced / untracked) LSN %v (%d) delta %d",
						cHash, xld.WALStart, xld.WALStart, lsnDelta)
					stats.reset()
					break // switch => continue loop
				}

				cid := binary.BigEndian.AppendUint64(nil, uint64(seq))
				cid = append(cid, cHash...)

				select {
				case commitHash <- cid:
				default: // don't block if the receiver has choked
					return fmt.Errorf("commit hash channel full")
				}

				lvl := log.DebugLevel
				if !bytes.Equal(cHash, zeroHash) {
					lvl = log.InfoLevel
				}
				logger.Logf(lvl, "Commit hash %x, seq %d, LSN %v (%d) delta %d",
					cHash, seq, xld.WALStart, xld.WALStart, lsnDelta)

				logger.Debug("wal commit stats", log.Uint("inserts", stats.inserts), log.Uint("updates", stats.updates),
					log.Uint("deletes", stats.deletes), log.Uint("truncates", stats.truncs))
				stats.reset()
				seq = -1 // next commit may be untracked, forget this one
			}

		default:
			logger.Debug("unknown message", log.String("type", string(msg.Data[0])))
		}
	}
}

type walStats struct {
	inserts uint64
	updates uint64
	deletes uint64
	truncs  uint64
}

func (ws *walStats) reset() {
	*ws = walStats{}
}

// decodeWALData decodes a wal data message given known relations, returning
// true if it was a commit message, or a non-negative seq value if it was a
// special update message on the internal sentry table
func decodeWALData(hasher hash.Hash, walData []byte, relations map[uint32]*pglogrepl.RelationMessageV2,
	inStream *bool, stats *walStats, okSchema func(schema string) bool, changesetWriter *changesetIoWriter) (bool, int64, error) {
	logicalMsg, err := parseV3(walData, *inStream)
	if err != nil {
		return false, 0, fmt.Errorf("parse logical replication message: %w", err)
	}

	var seq int64 = -1
	var done bool // set to true on receipt of a commit to signal the the end of a transaction

	switch logicalMsg := logicalMsg.(type) {
	case *pglogrepl.RelationMessageV2:
		logger.Debugf(" [msg] Relation: %d (%v.%v)", logicalMsg.RelationID,
			logicalMsg.Namespace, logicalMsg.RelationName)
		relations[logicalMsg.RelationID] = logicalMsg

	case *pglogrepl.BeginMessage:
		logger.Debugf(" [msg] Begin: LSN %v (%d)", logicalMsg.FinalLSN, uint64(logicalMsg.FinalLSN))
		// Indicates the beginning of a group of changes in a transaction. This
		// is only sent for committed transactions. You won't get any events
		// from rolled back transactions.

	case *pglogrepl.CommitMessage:
		logger.Debugf(" [msg] Commit: Commit LSN %v (%d), End LSN %v (%d)",
			logicalMsg.CommitLSN, uint64(logicalMsg.CommitLSN),
			logicalMsg.TransactionEndLSN, uint64(logicalMsg.TransactionEndLSN))

		done = true

	case *pglogrepl.InsertMessageV2:
		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			return false, 0, fmt.Errorf("insert: unknown relation ID %d", logicalMsg.RelationID)
		}

		relName := rel.Namespace + "." + rel.RelationName
		if !okSchema(rel.Namespace) {
			// logger.Debugf("ignoring update to relation %v", relName)
			break
		}

		err = changesetWriter.decodeInsert(logicalMsg, rel)
		if err != nil {
			return false, 0, fmt.Errorf("decode insert: %w", err)
		}

		insertData := encodeInsertMsg(relName, &logicalMsg.InsertMessage)
		// logger.Debugf("insertData %x", insertData)
		hasher.Write(insertData)

		logger.Debugf(" [msg] INSERT xid %d into rel %v.%v: %v", logicalMsg.Xid,
			rel.Namespace, rel.RelationName, &lazyValues{logicalMsg.Tuple.Columns, rel})

		stats.inserts++

	case *pglogrepl.UpdateMessageV2:
		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			return false, 0, fmt.Errorf("update: unknown relation ID %d", logicalMsg.RelationID)
		}

		// capture the seq value, before target schema filter
		if rel.Namespace == InternalSchemaName && rel.RelationName == sentryTableName {
			cols := logicalMsg.NewTuple.Columns
			if len(cols) != 1 {
				logger.Warnf("not one column in sentry table update (%d)", len(cols))
			} else {
				newSeq, err := cols[0].Int64()
				if err != nil {
					logger.Warnf("invalid sequence number in sentry table update: %v", err)
				} else {
					seq = newSeq
				}
			}
		}

		relName := rel.Namespace + "." + rel.RelationName
		if !okSchema(rel.Namespace) {
			// logger.Debugf("ignoring update to relation %v", relName)
			break
		}

		err = changesetWriter.decodeUpdate(logicalMsg, rel)
		if err != nil {
			return false, 0, fmt.Errorf("decode update: %w", err)
		}

		updateData := encodeUpdateMsg(relName, &logicalMsg.UpdateMessage)
		// logger.Debugf("updateData %x", updateData)
		hasher.Write(updateData)

		var oldValues *lazyValues
		if logicalMsg.OldTuple != nil { // seems to be only if primary key changes
			oldValues = &lazyValues{logicalMsg.OldTuple.Columns, rel}
		}
		logger.Debugf(" [msg] UPDATE rel %v.%v: %v => %v", rel.Namespace, rel.RelationName,
			oldValues, &lazyValues{logicalMsg.NewTuple.Columns, rel})

		stats.updates++

	case *pglogrepl.DeleteMessageV2:
		rel, ok := relations[logicalMsg.RelationID]
		if !ok {
			return false, 0, fmt.Errorf("delete: unknown relation ID %d", logicalMsg.RelationID)
		}

		relName := rel.Namespace + "." + rel.RelationName
		if !okSchema(rel.Namespace) {
			// logger.Debugf("ignoring update to relation %v", relName)
			break
		}

		err = changesetWriter.decodeDelete(logicalMsg, rel)
		if err != nil {
			return false, 0, fmt.Errorf("decode delete: %w", err)
		}

		deleteData := encodeDeleteMsg(relName, &logicalMsg.DeleteMessage)
		// logger.Debugf("deleteData %x", deleteData)
		hasher.Write(deleteData)

		logger.Debugf(" [msg] DELETE from rel %v.%v: %v", rel.Namespace, rel.RelationName,
			&lazyValues{logicalMsg.OldTuple.Columns, rel})

		stats.deletes++

	case *pglogrepl.TruncateMessageV2:
		rels := make(map[uint32]*pglogrepl.RelationMessageV2)
		for _, relID := range logicalMsg.RelationIDs {
			rel, ok := relations[relID]
			if !ok {
				logger.Warnf("unknown truncated relation ID %d", relID)
				continue
			}
			if okSchema(rel.Namespace) {
				rels[relID] = rel
				// relName := rel.Namespace + "." + rel.RelationName
			}
		}
		if len(rels) == 0 {
			logger.Debug("no relevant relations in truncate message")
			break
		}

		hasher.Write(encodeTruncateMsg(rels, &logicalMsg.TruncateMessage))
		stats.truncs++

	case *pglogrepl.TypeMessageV2:
		logger.Debugf("type message: %v %v %v", logicalMsg.Name, logicalMsg.Namespace, logicalMsg.DataType)
	case *pglogrepl.OriginMessage:
		logger.Debugf("origin message: %v %v", logicalMsg.Name, logicalMsg.CommitLSN)
	case *pglogrepl.LogicalDecodingMessageV2:
		logger.Debugf("logical decoding message: %q, %q, %d", logicalMsg.Prefix, logicalMsg.Content, logicalMsg.Xid)

	// prepared transaction messages
	case *BeginPrepareMessageV3:
		logger.Debugf(" [msg] BEGIN PREPARED TRANSACTION (id %v): Prepare LSN %v (%d), End LSN %v (%d) \n",
			logicalMsg.UserGID, logicalMsg.PrepareLSN, uint64(logicalMsg.PrepareLSN),
			logicalMsg.EndPrepareLSN, uint64(logicalMsg.EndPrepareLSN))
	case *PrepareMessageV3:
		logger.Debugf(" [msg] PREPARE TRANSACTION (id %v): Prepare LSN %v (%d), End LSN %v (%d) \n",
			logicalMsg.UserGID, logicalMsg.PrepareLSN, uint64(logicalMsg.PrepareLSN),
			logicalMsg.EndPrepareLSN, uint64(logicalMsg.EndPrepareLSN))

		// - BEGIN;
		// - mods: UPDATE / INSERT / DELETE
		// - PREPARE TRANSACTION 'uid';
		//	* msgs: Begin Prepared -> [update messages] -> Prepare (ready to commit)
		// - COMMIT PREPARED 'uid';
		//  * msgs: Commit Prepared (NO regular "Commit" message)
		done = true // there will be a commit or a rollback, but this is the end of the update stream

		err = changesetWriter.commit()
		if err != nil {
			return false, 0, fmt.Errorf("changeset commit error: %w", err)
		}

	case *CommitPreparedMessageV3:
		logger.Debugf(" [msg] COMMIT PREPARED TRANSACTION (id %v): Commit LSN %v (%d), End LSN %v (%d) \n",
			logicalMsg.UserGID, logicalMsg.CommitLSN, uint64(logicalMsg.CommitLSN),
			logicalMsg.EndCommitLSN, uint64(logicalMsg.EndCommitLSN))
		// done = true

	case *RollbackPreparedMessageV3:
		logger.Debugf(" [msg] ROLLBACK PREPARED TRANSACTION (id %v): Rollback LSN %v (%d), End LSN %v (%d) \n",
			logicalMsg.UserGID, logicalMsg.RollbackLSN, uint64(logicalMsg.RollbackLSN),
			logicalMsg.EndLSN, uint64(logicalMsg.EndLSN))

		hasher.Reset()
		changesetWriter.fail() // discard changeset

	// v2 Stream control messages.  Not expected for kwil
	case *pglogrepl.StreamStartMessageV2:
		*inStream = true
		logger.Warnf(" [msg] StreamStartMessageV2: xid %d, first segment? %d", logicalMsg.Xid, logicalMsg.FirstSegment)
	case *pglogrepl.StreamStopMessageV2:
		*inStream = false
		logger.Warnf(" [msg] StreamStopMessageV2")
	case *pglogrepl.StreamCommitMessageV2:
		logger.Warnf("Stream commit message: xid %d", logicalMsg.Xid)
	case *pglogrepl.StreamAbortMessageV2:
		logger.Warnf("Stream abort message: xid %d", logicalMsg.Xid)

	default:
		logger.Warnf("Unknown message type in pgoutput stream: %T", logicalMsg)
	}

	return done, seq, nil
}

// lazyValues is a fmt.Stringer used to lazily decode and print tuple column
// data (if required for a given log level).
type lazyValues struct {
	cols []*pglogrepl.TupleDataColumn
	rel  *pglogrepl.RelationMessageV2
}

func (lv *lazyValues) String() string {
	if lv == nil {
		return "<nil>"
	}
	values, err := tuplColVals(lv.cols, lv.rel)
	if err != nil {
		logger.Warn("column value decoding", log.String("error", err.Error()))
		return "<invalid>" // may not be logged at all by caller
	}
	return fmt.Sprintf("%v", values) // alt: json.Encode to make it slightly prettier
}

func init() {
	typeMap = pgtype.NewMap()
}

var typeMap *pgtype.Map

func decodeTextColumnData(data []byte, dataType uint32) (interface{}, error) {
	if dt, ok := typeMap.TypeForOID(dataType); ok {
		return dt.Codec.DecodeValue(typeMap, dataType, pgtype.TextFormatCode, data)
	}
	return string(data), nil
}

func tuplColVals(cols []*pglogrepl.TupleDataColumn, rel *pglogrepl.RelationMessageV2) (map[string]any, error) {
	values := map[string]any{}
	for idx, col := range cols {
		colName := rel.Columns[idx].Name
		switch col.DataType {
		case 'n': // null
			values[colName] = nil
		case 'u': // unchanged toast
			// This TOAST value was not changed. TOAST values are not stored
			// in the tuple, and logical replication doesn't want to spend a
			// disk read to fetch its value for you.
		case 't': //text
			val, err := decodeTextColumnData(col.Data, rel.Columns[idx].DataType)
			if err != nil {
				return nil, fmt.Errorf("error decoding column data: %w", err)
			}
			values[colName] = val
		}
	}
	return values, nil
}

// The following encodings of the insert/update/delete/truncate messages
// directly affect the commit hash and are thus consensus code. Edit with care.

var pgIntCoder = binary.BigEndian

func encodeTupleData(td *pglogrepl.TupleData) []byte {
	if td == nil {
		return []byte{0}
	}
	var data []byte
	data = pgIntCoder.AppendUint16(data, td.ColumnNum)
	for _, col := range td.Columns {
		data = append(data, col.DataType)

		switch col.DataType {
		case pglogrepl.TupleDataTypeText, pglogrepl.TupleDataTypeBinary:
			pgIntCoder.AppendUint32(data, col.Length) // len(col.Data)
			data = append(data, col.Data...)
		case pglogrepl.TupleDataTypeNull, pglogrepl.TupleDataTypeToast:
		}
	}
	return data
}

func encodeInsertMsg(relName string, im *pglogrepl.InsertMessage) []byte {
	data := []byte(relName) // RelationID is dependent on the deployment
	return append(data, encodeTupleData(im.Tuple)...)
}

func encodeUpdateMsg(relName string, um *pglogrepl.UpdateMessage) []byte {
	data := []byte(relName) // RelationID is dependent on the deployment
	data = append(data, um.OldTupleType)
	data = append(data, encodeTupleData(um.OldTuple)...)
	return append(data, encodeTupleData(um.NewTuple)...)
}

func encodeDeleteMsg(relName string, um *pglogrepl.DeleteMessage) []byte {
	data := []byte(relName) // RelationID is dependent on the deployment
	data = append(data, um.OldTupleType)
	return append(data, encodeTupleData(um.OldTuple)...)
}

func encodeTruncateMsg(rels map[uint32]*pglogrepl.RelationMessageV2, tm *pglogrepl.TruncateMessage) []byte {
	var buf bytes.Buffer
	buf.WriteByte(tm.Option)
	for _, rid := range tm.RelationIDs {
		rel, ok := rels[rid]
		if !ok {
			continue // not a relevant relation
		}
		relName := rel.Namespace + "." + rel.RelationName
		buf.WriteString(relName)
	}
	return buf.Bytes()
}
