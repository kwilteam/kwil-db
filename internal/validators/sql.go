// Package validators contain the "validator store" for persistent storage of
// active validators, and candidate validators along with approval tx records.
// This facilitates reloading validator state, which includes active votes.
//
// When a prospective validator submits a join tx, they are inserted into the
// validators table with a power of 0. When current validators submit an
// approve tx, the
package validators

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/sql"
)

var errUnknownValidator = errors.New("unknown validator")

// valStoreVersion is the current schema version.
var valStoreVersion = 0

// store:
// - current validator set
// - active approvals/votes
// - removal proposals
const (
	schemaName      = `kwild_vals`
	sqlCreateSchema = `CREATE SCHEMA IF NOT EXISTS ` + schemaName

	sqlInitTables = `
	CREATE TABLE IF NOT EXISTS ` + schemaName + `.validators (
		pubkey BYTEA PRIMARY KEY,
		power BIGINT
	);

	-- removals contains all validator removal proposals / votes from a
	-- given remover validator targeting another validator.
	-- If the targeted validator is ultimately removed or voluntarily leaves
	-- the validator set, all relevant removal request should be removed.
	CREATE TABLE IF NOT EXISTS ` + schemaName + `.removals (
		remover BYTEA REFERENCES ` + schemaName + `.validators (pubkey) ON DELETE CASCADE,
		target BYTEA REFERENCES ` + schemaName + `.validators (pubkey) ON DELETE CASCADE,
		PRIMARY KEY (remover, target)
	);

	CREATE TABLE IF NOT EXISTS ` + schemaName + `.join_reqs (
		candidate BYTEA PRIMARY KEY,
		power_wanted BIGINT,
		expiresAt BIGINT
	);

	CREATE TABLE IF NOT EXISTS ` + schemaName + `.joins_board (
		candidate BYTEA REFERENCES ` + schemaName + `.join_reqs (candidate) ON DELETE CASCADE,  -- not in the validators table yet
		validator BYTEA REFERENCES ` + schemaName + `.validators (pubkey) ON DELETE CASCADE,
		approval BIGINT,
		PRIMARY KEY (candidate, validator)
	);`

	// joins_board give us the board of validators (approvers) for a given join
	// request which is needed to resume vote handling. The validators for a
	// candidate are determined at the time the join request is created.

	sqlSetApproval = `UPDATE ` + schemaName + `.joins_board SET approval = $1
		WHERE validator = $2 AND candidate = $3`

	sqlActiveValidators = `SELECT pubkey, power FROM ` + schemaName + `.validators WHERE power > 0`

	// get the rows: candidate, power - separate query for scan prealloc
	sqlGetOngoingVotes = `SELECT candidate, power_wanted, expiresAt FROM ` + schemaName + `.join_reqs;`
	// a validator "join" request for a candidate may receive votes from a
	// specific set of existing validators, calling this the board of
	// validators.
	sqlVoteStatus = `SELECT validator, approval
		FROM ` + schemaName + `.joins_board
		WHERE candidate = $1`

	sqlEligibleApprove = `SELECT 1 FROM ` + schemaName + `.joins_board
		WHERE candidate = $1 AND validator = $2`

	sqlDeleteAllValidators = "DELETE FROM " + schemaName + ".validators;"
	sqlDeleteAllJoins      = "DELETE FROM " + schemaName + ".join_reqs;"

	// NOTE: if re-adding a validator, this will hit the UNIQUE constraint on
	// pubkey. We may want to keep validators in the table with power 0 on leave
	// or punish, so we perform an upsert to be safe.
	sqlNewValidator = `INSERT INTO ` + schemaName + `.validators (pubkey, power) VALUES ($1, $2)
		ON CONFLICT (pubkey) DO UPDATE SET power = $2` // NOTE: pg requires cols or constraint name. update parser?
	sqlDeleteValidator      = `DELETE FROM ` + schemaName + `.validators WHERE pubkey = $1;`
	sqlUpdateValidatorPower = `UPDATE ` + schemaName + `.validators SET power = $1
		WHERE pubkey = $2`
	sqlGetValidatorPower = `SELECT power FROM ` + schemaName + `.validators WHERE pubkey = $1`

	sqlNewJoinReq = `INSERT INTO ` + schemaName + `.join_reqs (candidate, power_wanted, expiresAt)
		VALUES ($1, $2, $3)`
	sqlDeleteJoinReq = `DELETE FROM ` + schemaName + `.join_reqs WHERE candidate = $1;` // cascades to joins_board

	sqlAddToJoinBoard = `INSERT INTO ` + schemaName + `.joins_board (candidate, validator, approval)
		VALUES ($1, $2, $3)`

	sqlListAllRemovals    = `SELECT target, remover FROM ` + schemaName + `.removals`
	sqlListTargetRemovals = `SELECT remover FROM ` + schemaName + `.removals WHERE target = $1`
	sqlAddRemoval         = `INSERT INTO ` + schemaName + `.removals (remover, target) VALUES ($1, $2)`
	sqlDeleteRemoval      = `DELETE FROM ` + schemaName + `.removals WHERE remover = $1 AND target = $2`
	sqlDeleteRemovals     = `DELETE FROM ` + schemaName + `.removals WHERE target = $1`

	// Schema version table queries
	sqlInitVersionTable = `CREATE TABLE IF NOT EXISTS ` + schemaName + `.schema_version (
		version INT4 NOT NULL
    );`

	sqlInitVersionRow = `INSERT INTO ` + schemaName + `.schema_version (version) VALUES ($1);`

	// sqlUpdateVersion = `UPDATE ` + schemaName + `.schema_version SET version = $1;`

	sqlGetVersion = `SELECT version FROM ` + schemaName + `.schema_version;`
)

// func (vs *validatorStore) updateCurrentVersion(ctx context.Context, version int) error {
// 	_, err := vs.db.Execute(ctx, sqlUpdateVersion, version)
// 	if err != nil {
// 		return fmt.Errorf("failed to update schema version: %w", err)
// 	}
// 	return nil
// }

func (vs *validatorStore) currentVersion(ctx context.Context, db sql.DB) (int, error) {
	// we need a tx because a postgres query against non-existent relation will fail the whole tx
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) // since we are reading, we can just rollback

	results, err := tx.Execute(ctx, sqlGetVersion)
	if err != nil {
		return 0, err
	}
	if len(results.Rows) == 0 {
		return 0, nil
	}
	if len(results.Rows) > 1 {
		return 0, fmt.Errorf("expected 1 row, got %d", len(results.Rows))
	}

	// rows[0][0] == version

	version, ok := results.Rows[0][0].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid version value (%T)", version)
	}

	return int(version), nil
}

func getTableInits() []string {
	inits := strings.Split(sqlInitTables, ";")
	return inits[:len(inits)-1]
}

// initTables initializes the validator store tables. This is not an upgrade
// action and is only used on a fresh DB being created at the latest version.
func (vs *validatorStore) initTables(ctx context.Context, db sql.DB) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Execute(ctx, sqlCreateSchema); err != nil {
		return err
	}

	inits := getTableInits()
	for _, init := range inits {
		if _, err := tx.Execute(ctx, init); err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
		}
	}

	if _, err := tx.Execute(ctx, sqlInitVersionTable); err != nil {
		return fmt.Errorf("failed to initialize schema version table: %w", err)
	}

	_, err = tx.Execute(ctx, sqlInitVersionRow, valStoreVersion)
	if err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}

	return tx.Commit(ctx)
}

func (vs *validatorStore) startJoinRequest(ctx context.Context, tx sql.DB, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error {
	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	// Insert into join_reqs.
	_, err = sp.Execute(ctx, sqlNewJoinReq, joiner, power, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert new join request: %w", err)
	}

	// Insert all approvers into joins_board.
	//  TODO: Maybe prepare a statement since we execute in a loop.
	// newJoinBoardStmt, err := vs.db.Prepare(sqlAddToJoinBoard)
	// if err != nil {
	// 	newJoinBoardStmt.Close()
	// 	return fmt.Errorf("failed to prepare get account statement: %w", err)
	// }
	for i := range approvers {
		_, err = sp.Execute(ctx, sqlAddToJoinBoard, joiner, approvers[i], 0)
		if err != nil {
			return fmt.Errorf("failed to insert new join request: %w", err)
		}
	}
	return sp.Commit(ctx)
}

func (vs *validatorStore) addApproval(ctx context.Context, tx sql.DB, joiner, approver []byte) error {
	// We could just YOLO update, potentially updating zero rows if there's no
	// join request for this candidate or if approver is not an eligible voting
	// validator, but let's go the extra mile.
	res, err := tx.Execute(ctx, sqlEligibleApprove, joiner, approver)
	if err != nil {
		return err
	}
	if len(res.Rows) != 1 {
		return fmt.Errorf("%d eligible join requests to approve", len(res.Rows))
	}

	// Update the approval column of join_board row.
	_, err = tx.Execute(ctx, sqlSetApproval, 1, approver, joiner)
	return err
}

func (vs *validatorStore) addRemoval(ctx context.Context, tx sql.DB, target, validator []byte) error {
	_, err := tx.Execute(ctx, sqlAddRemoval, validator, target)
	return err
}

// deleteRemoval and deleteRemovals should not be required with ON DELETE
// CASCADE on both the remover and target columns of the removals table...
func (vs *validatorStore) deleteRemoval(ctx context.Context, tx sql.DB, target, validator []byte) error {
	_, err := tx.Execute(ctx, sqlDeleteRemoval, validator, target)
	return err
}

func (vs *validatorStore) deleteJoinRequest(ctx context.Context, tx sql.DB, joiner []byte) error {
	_, err := tx.Execute(ctx, sqlDeleteJoinReq, joiner)
	return err
}

func (vs *validatorStore) addValidator(ctx context.Context, tx sql.DB, joiner []byte, power int64) error {
	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	// Only permit this for first time validators (unknown) or validators with
	// power zero (not active, but in our tables).
	power0, err := vs.validatorPower(ctx, sp, joiner)
	if err != nil && !errors.Is(err, errUnknownValidator) {
		return err
	}
	if power0 > 0 {
		return errors.New("validator with power already exists")
	}
	// Either a new validator, or we are doing a power upsert.
	_, err = sp.Execute(ctx, sqlNewValidator, joiner, power)
	if err != nil {
		return fmt.Errorf("failed to add validator: %w", err)
	}
	err = vs.deleteJoinRequest(ctx, sp, joiner)
	if err != nil {
		return fmt.Errorf("failed to delete join request: %w", err)
	}

	return sp.Commit(ctx)
}

func (vs *validatorStore) removeValidator(ctx context.Context, tx sql.DB, validator []byte) error {
	sp, err := tx.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer sp.Rollback(ctx)

	_, err = sp.Execute(ctx, sqlDeleteRemovals, validator)
	if err != nil {
		return fmt.Errorf("failed to delete removals: %w", err)
	}
	_, err = sp.Execute(ctx, sqlDeleteValidator, validator)
	if err != nil {
		return fmt.Errorf("failed to remove validator: %w", err)
	}
	return sp.Commit(ctx)
}

func (vs *validatorStore) updateValidatorPower(ctx context.Context, tx sql.DB, validator []byte, power int64) error {
	_, err := tx.Execute(ctx, sqlUpdateValidatorPower, power, validator)
	if err != nil {
		return fmt.Errorf("failed to update validator power: %w", err)
	}
	return nil
}

func (vs *validatorStore) init(ctx context.Context, tx sql.DB, vals []*Validator) error {
	_, err := tx.Execute(ctx, sqlDeleteAllValidators)
	if err != nil {
		return fmt.Errorf("failed to delete all previous validators: %w", err)
	}
	_, err = tx.Execute(ctx, sqlDeleteAllJoins)
	if err != nil {
		return fmt.Errorf("failed to delete all previous join requests: %w", err)
	}

	for _, vi := range vals {
		_, err = tx.Execute(ctx, sqlNewValidator, vi.PubKey, vi.Power)
		if err != nil {
			return fmt.Errorf("failed to insert validator: %w", err)
		}
	}

	return nil
}

func (vs *validatorStore) validatorPower(ctx context.Context, tx sql.DB, validator []byte) (int64, error) {
	results, err := tx.Execute(ctx, sqlGetValidatorPower, validator)
	if err != nil {
		return 0, err
	}
	if len(results.Rows) == 0 {
		return 0, errUnknownValidator
	}
	if len(results.Rows[0]) != 1 {
		return 0, fmt.Errorf("expected 1 column getting validator power. this is an internal bug")
	}

	power, ok := results.Rows[0][0].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid power value (%T)", power)
	}

	return power, nil
}

func (vs *validatorStore) currentValidators(ctx context.Context, tx sql.DB) ([]*Validator, error) {
	results, err := tx.Execute(ctx, sqlActiveValidators)
	if err != nil {
		return nil, err
	}
	if len(results.Rows) == 0 {
		return nil, nil // no validators, ok, skip the slice alloc
	}

	vals := make([]*Validator, len(results.Rows))
	for i, row := range results.Rows {
		if len(row) != 2 {
			return nil, fmt.Errorf("expected 2 columns getting validators. this is an internal bug")
		}

		// row[0] is the pubkey, row[1] is the power
		pub, ok := row[0].([]byte)
		if !ok {
			return nil, errors.New("no pubkey in validator record")
		}

		power, ok := row[1].(int64)
		if !ok {
			return nil, errors.New("no power in validator record")
		}
		vals[i] = &Validator{
			PubKey: pub,
			Power:  power,
		}
	}
	return vals, nil
}

func (vs *validatorStore) allActiveRemoveReqs(ctx context.Context, tx sql.DB) ([]*ValidatorRemoveProposal, error) {
	results, err := tx.Execute(ctx, sqlListAllRemovals)
	if err != nil {
		return nil, err
	}

	removals := make([]*ValidatorRemoveProposal, len(results.Rows))
	for i, row := range results.Rows {
		// row[0] is the target, row[1] is the remover
		target, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid target pubkey value (%T)", target)
		}

		remover, ok := row[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid remover pubkey value (%T)", remover)
		}

		removals[i] = &ValidatorRemoveProposal{
			Target:  target,
			Remover: remover,
		}
	}

	return removals, nil
}

func (vs *validatorStore) loadJoinVotes(ctx context.Context, tx sql.DB, jr *JoinRequest) error {
	results, err := tx.Execute(ctx, sqlVoteStatus, jr.Candidate)
	if err != nil {
		return err
	}

	candidateVotes, err := voteStatusFromRecords(results)
	if err != nil {
		return err
	}
	for _, vi := range candidateVotes {
		jr.Board = append(jr.Board, vi.pubkey)
		jr.Approved = append(jr.Approved, vi.approval > 0)
	}
	return nil
}

func (vs *validatorStore) allActiveJoinReqs(ctx context.Context, tx sql.DB) ([]*JoinRequest, error) {
	results, err := tx.Execute(ctx, sqlGetOngoingVotes)
	if err != nil {
		return nil, err
	}
	if len(results.Rows) == 0 {
		return nil, nil
	}
	candidates, err := activeVotesFromRecords(results)
	if err != nil {
		return nil, err
	}

	allJoins := make([]*JoinRequest, len(candidates))

	for i, cv := range candidates {
		jr := &JoinRequest{
			Candidate: cv.pubkey,
			Power:     cv.pwr,
			ExpiresAt: cv.expiresAt,
		}
		err = vs.loadJoinVotes(ctx, tx, jr)
		if err != nil {
			return nil, err
		}
		allJoins[i] = jr
	}

	return allJoins, nil
}

type candidate struct {
	pubkey    []byte
	pwr       int64
	expiresAt int64
}

func activeVotesFromRecords(results *sql.ResultSet) ([]*candidate, error) {
	vals := make([]*candidate, len(results.Rows))
	for i, row := range results.Rows {
		// row[0] is the pubkey, row[1] is the power wanted, row[2] is the expiresAt
		if len(row) != 3 {
			return nil, fmt.Errorf("expected 3 columns getting join requests. this is an internal bug")
		}

		pubkey, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid pubkey value (%T)", row[0])
		}

		power, ok := row[1].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid power value (%T)", row[1])
		}

		expiresAt, ok := row[2].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid expiresAt value (%T)", row[2])
		}

		vals[i] = &candidate{
			pubkey:    pubkey,
			pwr:       power,
			expiresAt: expiresAt,
		}

	}
	return vals, nil
}

type approver struct {
	pubkey   []byte
	approval int64
}

func voteStatusFromRecords(results *sql.ResultSet) ([]*approver, error) {
	if len(results.Rows) == 0 {
		return nil, errors.New("no results")
	}

	board := make([]*approver, len(results.Rows))
	for i, row := range results.Rows {
		// row[0] is the validator, row[1] is the approval
		if len(row) != 2 {
			return nil, fmt.Errorf("expected 2 columns getting join requests. this is an internal bug")
		}

		pubkey, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid pubkey value (%T)", row[0])
		}

		approval, ok := row[1].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid approval value (%T)", row[1])
		}

		board[i] = &approver{
			pubkey:   pubkey,
			approval: approval,
		}
	}
	return board, nil
}

// preparedStatements are used with Execute when performing read-write queries,
// or to see uncommitted changes.
/* unused presently
type preparedStatements struct {
	newJoinReq sql.Statement
}

func (p *preparedStatements) Close() error {
	return p.newJoinReq.Close()
}

func (vs *validatorStore) prepareStatements() error {
	if vs.stmts == nil {
		vs.stmts = &preparedStatements{}
	}

	stmt, err := vs.db.Prepare(sqlNewJoinReq)
	if err != nil {
		return fmt.Errorf("failed to prepare get account statement: %w", err)
	}
	vs.stmts.newJoinReq = stmt

	return nil
}
*/
