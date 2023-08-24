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
)

// store:
// - current validator set
// - active approvals/votes
const (
	sqlInitTables = `
	CREATE TABLE IF NOT EXISTS validators (
		pubkey BLOB PRIMARY KEY,
		power INTEGER
		) WITHOUT ROWID, STRICT;

	CREATE TABLE IF NOT EXISTS join_reqs (
		candidate BLOB PRIMARY KEY,
		power_wanted INTEGER
	) WITHOUT ROWID, STRICT;

	CREATE TABLE IF NOT EXISTS joins_board (
		candidate BLOB REFERENCES join_reqs (candidate) ON DELETE CASCADE,  -- not in the validators table yet
		validator BLOB REFERENCES validators (pubkey) ON DELETE CASCADE,
		approval INTEGER,
		PRIMARY KEY (candidate, validator)
	) WITHOUT ROWID, STRICT;`

	// joins_board give us the board of validators (approvers) for a given join
	// request which is needed to resume vote handling. The validators for a
	// candidate are determined at the time the join request is created.

	sqlSetApproval = `UPDATE joins_board SET approval = $approval
		WHERE validator = $validator AND candidate = $candidate`

	sqlActiveValidators = `SELECT pubkey, power FROM validators
		WHERE power > 0 COLLATE NOCASE`

	// get the rows: candidate, power - separate query for scan prealloc
	sqlGetOngoingVotes = "SELECT candidate, power_wanted FROM join_reqs;"
	// a validator "join" request for a candidate may receive votes from a
	// specific set of existing validators, calling this the board of
	// validators.
	sqlVoteStatus = `SELECT validator, approval
		FROM joins_board
		WHERE candidate = $candidate`

	sqlEligibleApprove = `SELECT 1 FROM joins_board
		WHERE candidate = $candidate AND validator = $validator`

	sqlDeleteAllValidators = "DELETE FROM validators;"
	sqlDeleteAllJoins      = "DELETE FROM join_reqs;"

	sqlNewValidator         = "INSERT INTO validators (pubkey, power) VALUES ($pubkey, $power)"
	sqlDeleteValidator      = "DELETE FROM validators WHERE pubkey = $pubkey;"
	sqlUpdateValidatorPower = `UPDATE validators SET power = $power
		WHERE pubkey = $pubkey`

	sqlNewJoinReq = `INSERT INTO join_reqs (candidate, power_wanted)
		VALUES ($candidate, $power_wanted)`
	sqlDeleteJoinReq = "DELETE FROM join_reqs WHERE candidate = $candidate;" // cascades to joins_board

	sqlAddToJoinBoard = `INSERT INTO joins_board (candidate, validator, approval)
		VALUES ($candidate, $validator, $approval)`
)

// -- CREATE TABLE IF NOT EXISTS validator_approvals (
// -- 	validator_id INTEGER REFERENCES validators (id) ON DELETE CASCADE,
// -- 	approval_id INTEGER REFERENCES approvals (id) ON DELETE CASCADE,
// -- 	unique (validator_id, approval_id)
// -- 	);

// -- CREATE TABLE IF NOT EXISTS approvals (
// -- 	id INTEGER PRIMARY KEY AUTOINCREMENT,
// -- 	voter TEXT NOT NULL   -- the validator that approved
// -- 	);

func getTableInits() []string {
	inits := strings.Split(sqlInitTables, ";")
	return inits[:len(inits)-1]
}

func (vs *validatorStore) initTables(ctx context.Context) error {
	inits := getTableInits()

	for _, init := range inits {
		err := vs.db.Execute(ctx, init, nil)
		if err != nil {
			return fmt.Errorf("failed to initialize tables: %w", err)
		}
	}

	return nil
}

func (vs *validatorStore) startJoinRequest(ctx context.Context, joiner []byte, approvers [][]byte, power int64) error {
	// Insert into join_reqs.
	err := vs.db.Execute(ctx, sqlNewJoinReq, map[string]any{
		"$candidate":    joiner,
		"$power_wanted": power,
	})
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
		err = vs.db.Execute(ctx, sqlAddToJoinBoard, map[string]any{
			"$candidate": joiner,
			"$validator": approvers[i],
			"$approval":  0,
		})
		if err != nil {
			return fmt.Errorf("failed to insert new join request: %w", err)
		}
	}
	return nil
}

func (vs *validatorStore) addApproval(ctx context.Context, joiner, approver []byte) error {
	// We could just YOLO update, potentially updating zero rows if there's no
	// join request for this candidate or if approver is not an eligible voting
	// validator, but let's go the extra mile.
	res, err := vs.db.Query(ctx, sqlEligibleApprove, map[string]any{
		"$candidate": joiner,
		"$validator": approver,
	})
	if err != nil {
		return err
	}
	if len(res) != 1 {
		return fmt.Errorf("%d eligible join requests to approve", len(res))
	}

	// Update the approval column of join_board row.
	return vs.db.Execute(ctx, sqlSetApproval, map[string]any{
		"$approval":  1,
		"$validator": approver,
		"$candidate": joiner,
	})
}

func (vs *validatorStore) addValidator(ctx context.Context, joiner []byte, power int64) error {
	err := vs.db.Execute(ctx, sqlNewValidator, map[string]any{
		"$pubkey": joiner,
		"$power":  power,
	})
	if err != nil {
		return fmt.Errorf("failed to add validator: %w", err)
	}
	err = vs.db.Execute(ctx, sqlDeleteJoinReq, map[string]any{
		"$candidate": joiner,
	})
	if err != nil {
		return fmt.Errorf("failed to remove join request: %w", err)
	}

	return nil
}

func (vs *validatorStore) removeValidator(ctx context.Context, validator []byte) error {
	err := vs.db.Execute(ctx, sqlDeleteValidator, map[string]any{
		"$pubkey": validator,
	})
	if err != nil {
		return fmt.Errorf("failed to remove validator: %w", err)
	}
	return nil
}

func (vs *validatorStore) updateValidatorPower(ctx context.Context, validator []byte, power int64) error {
	err := vs.db.Execute(ctx, sqlUpdateValidatorPower, map[string]any{
		"$power":  power,
		"$pubkey": validator,
	})
	if err != nil {
		return fmt.Errorf("failed to update validator power: %w", err)
	}
	return nil
}

func (vs *validatorStore) init(ctx context.Context, vals []*Validator) error {
	err := vs.db.Execute(ctx, sqlDeleteAllValidators, map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to delete all previous validators: %w", err)
	}
	err = vs.db.Execute(ctx, sqlDeleteAllJoins, map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to delete all previous join requests: %w", err)
	}

	for _, vi := range vals {
		err = vs.db.Execute(ctx, sqlNewValidator, map[string]any{
			"$pubkey": vi.PubKey,
			"$power":  vi.Power,
		})
		if err != nil {
			return fmt.Errorf("failed to insert validator: %w", err)
		}
	}

	return nil
}

func (vs *validatorStore) currentValidators(ctx context.Context) ([]*Validator, error) {
	results, err := vs.db.Query(ctx, sqlActiveValidators, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil // no validators, ok, skip the slice alloc
	}

	vals := make([]*Validator, len(results))
	for i, res := range results {
		pki, ok := res["pubkey"]
		if !ok {
			return nil, errors.New("no pubkey in validator record")
		}
		pubkey, ok := pki.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid pubkey value (%T)", pki)
		}
		pwri, ok := res["power"]
		if !ok {
			return nil, errors.New("no power in validator record")
		}
		power, ok := pwri.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid power value (%T)", pwri)
		}
		vals[i] = &Validator{
			PubKey: pubkey,
			Power:  power,
		}
	}
	return vals, nil
}

func (vs *validatorStore) loadJoinVotes(ctx context.Context, jr *JoinRequest) error {
	results, err := vs.db.Query(ctx, sqlVoteStatus, map[string]interface{}{
		"$candidate": jr.Candidate,
	})
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

func (vs *validatorStore) allActiveJoinReqs(ctx context.Context) ([]*JoinRequest, error) {
	results, err := vs.db.Query(ctx, sqlGetOngoingVotes, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
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
		}
		err = vs.loadJoinVotes(ctx, jr)
		if err != nil {
			return nil, err
		}
		allJoins[i] = jr
	}

	return allJoins, nil
}

type candidate struct {
	pubkey []byte
	pwr    int64
}

func activeVotesFromRecords(results []map[string]interface{}) ([]*candidate, error) {
	vals := make([]*candidate, len(results))
	for i, res := range results {
		pki, ok := res["candidate"]
		if !ok {
			return nil, errors.New("no candidate in join_reqs record")
		}
		pubkey, ok := pki.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid pubkey value (%T)", pki)
		}
		pwri, ok := res["power_wanted"]
		if !ok {
			return nil, errors.New("no power_wanted in join_reqs record")
		}
		power, ok := pwri.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid power value (%T)", pwri)
		}
		vals[i] = &candidate{
			pubkey: pubkey,
			pwr:    power,
		}
	}
	return vals, nil
}

type approver struct {
	pubkey   []byte
	approval int64
}

func voteStatusFromRecords(results []map[string]interface{}) ([]*approver, error) {
	if len(results) == 0 {
		return nil, errors.New("no results")
	}

	board := make([]*approver, len(results))
	for i, res := range results {
		pki, ok := res["validator"]
		if !ok {
			return nil, errors.New("no pubkey in joins_board record")
		}
		pubkey, ok := pki.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid pubkey value (%T)", pki)
		}
		appr, ok := res["approval"]
		if !ok {
			return nil, errors.New("no approval in joins_board record")
		}
		approval, ok := appr.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid approval value (%T)", appr)
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