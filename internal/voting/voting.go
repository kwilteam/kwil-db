package voting

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// NewVoteProcessor creates a new vote processor.
// It initializes the database with the required tables.
// The threshold is the percentThreshold of votes required to approve a resolution
// It must be an integer between 0 and 1000000.  This defines the percentage
func NewVoteProcessor(ctx context.Context, db sql.DB, accounts AccountStore,
	engine Engine, threshold int64, logger log.Logger) (*VoteProcessor, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	initStmts := []string{createVotingSchema, tableResolutionTypes, tableResolutions,
		resolutionsTypeIndex, tableProcessed, tableVoters, tableVotes} // order important

	for _, stmt := range initStmts {
		_, err := tx.Execute(ctx, stmt)
		if err != nil {
			return nil, err
		}
	}

	for name := range registeredPayloads {
		uuid := types.NewUUIDV5([]byte(name))
		_, err := tx.Execute(ctx, createResolutionType, uuid[:], name)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &VoteProcessor{
		percentThreshold: threshold,
		accounts:         accounts,
		engine:           engine,
		logger:           logger,
	}, nil
}

// VoteProcessor manages in-process votes, and manages how they get processed.
// It is responsible for tracking the relative power of voters, and for
// expiring resolutions.
type VoteProcessor struct {
	// percentThreshold is the percentThreshold of votes required to approve a resolution
	// it must be a number between 0 and 1000000.  This defines the percentage
	// of total voting power required to approve a resolution (e.g. 500000 = 50%)
	percentThreshold int64

	accounts AccountStore
	engine   Engine

	logger log.Logger
}

// ResolutionVoteInfo is a struct that contains information about the status of a resolution
type ResolutionVoteInfo struct {
	Resolution

	// ApprovedPower is the aggregate amount of power that has approved
	// the vote.
	ApprovedPower int64
}

// Resolution is a struct that contains information about a resolution
type Resolution struct {
	// ID is the unique identifier of the vote
	ID types.UUID

	// Type is a string to identify the type of a vote.
	// This can be nil, if no body has been added yet.
	Type string

	// Body is the actual vote
	// This can be nil, if the vote was began before a
	// body was added.
	Body []byte

	// Expiration is the blockheight at which the vote expires
	Expiration int64
}

// Approve approves a resolution from a voter.
// If the resolution does not yet exist, it will be created.
// If created, it will not be given a body, and can be given a body later.
// If the resolution already exists, it will simply track that the voter
// has approved the resolution, and will not change the body or expiration.
// If the voter does not exist, an error will be returned.
// If the voter has already approved the resolution, no error will be returned.
// If the resolution has already been processed, no error will be returned.
func (v *VoteProcessor) Approve(ctx context.Context, db sql.DB, resolutionID types.UUID, expiration int64, from []byte) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	alreadyProcessed, err := v.IsProcessed(ctx, tx, resolutionID)
	if err != nil {
		return err
	}
	if alreadyProcessed {
		return nil
	}

	// we need to ensure that the resolution ID exists
	_, err = tx.Execute(ctx, resolutionIDExists, resolutionID[:], expiration)
	if err != nil {
		return err
	}

	userId := types.NewUUIDV5(from)

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, the following will error
	_, err = tx.Execute(ctx, addVote, resolutionID[:], userId[:])
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreateResolution creates a vote, by submitting a body of a vote, a topic
// and an expiration.  The expiration should be a blockheight.
// If the resolution already exists, it will not be changed.
// If the resolution was already processed, nothing will happen.
func (v *VoteProcessor) CreateResolution(ctx context.Context, db sql.DB, event *types.VotableEvent, expiration int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	alreadyProcessed, err := v.IsProcessed(ctx, tx, event.ID())
	if err != nil {
		return err
	}
	if alreadyProcessed {
		return nil
	}

	// ensure that the category exists
	_, ok := registeredPayloads[event.Type]
	if !ok {
		return fmt.Errorf("payload not registered for type %s", event.Type)
	}

	id := event.ID()

	_, err = tx.Execute(ctx, upsertResolution, id[:], event.Body, event.Type, expiration)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Expire expires all votes at or before a given blockheight.
// All expired votes will be removed from the system.
func (v *VoteProcessor) Expire(ctx context.Context, db sql.DB, blockheight int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, expireResolutions, blockheight)
	if err != nil {
		return err
	}

	if len(res.Columns) != 1 {
		// this should never happen, just for safety
		return fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	for _, row := range res.Rows {
		id, ok := row[0].([]byte)
		if !ok {
			return fmt.Errorf("invalid type for id (%T). this is an internal bug", row[0])
		}

		_, err = tx.Execute(ctx, markProcessed, id)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetResolutionVoteInfo gets a resolution, identified by the ID.
// It does not read uncommitted data.
func (v *VoteProcessor) GetResolutionVoteInfo(ctx context.Context, db sql.DB, id types.UUID) (info *ResolutionVoteInfo, err error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, getResolutionVoteInfo, id[:]) // TODO: register our UUID with scanner's type map?
	if err != nil {
		return nil, err
	}

	// if no rows, then the resolution may still exist, but does not have a body
	if len(res.Rows) == 0 {
		res, err = tx.Execute(ctx, getUnfilledResolutionVoteInfo, id[:])
		if err != nil {
			return nil, err
		}

		if len(res.Rows) == 0 {
			return nil, ErrResolutionNotFound
		}

		// res.ReturnedColumns[0] == expiration
		// res.ReturnedColumns[1] == approved_power
		if len(res.Rows[0]) != 2 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
		}

		expIface := res.Rows[0][0]
		expiration, ok := expIface.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid type for expiration (%T). this is an internal bug", expIface)
		}
		powerIface := res.Rows[0][1]
		approvedPower, ok := sql.Int64(powerIface)
		if !ok {
			return nil, fmt.Errorf("invalid type for approved power (%T). this is an internal bug", powerIface)
		}

		return &ResolutionVoteInfo{
			Resolution: Resolution{
				ID:         id,
				Expiration: expiration,
			},
			ApprovedPower: approvedPower,
		}, nil
	}

	// res.ReturnedColumns[0] == id
	// res.ReturnedColumns[1] == body (can be nil)
	// res.ReturnedColumns[2] == type (can be nil)
	// res.ReturnedColumns[3] == expiration
	// res.ReturnedColumns[4] == approved_power
	if len(res.Rows[0]) != 5 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	resolution, err := getResolutionFromRow(res.Rows[0])
	if err != nil {
		return nil, fmt.Errorf("getResolutionFromRow: %w", err)
	}
	powerIface := res.Rows[0][4]
	approvedPower, ok := sql.Int64(powerIface)
	if !ok {
		return nil, fmt.Errorf("invalid type for approved power (%T). this is an internal bug", powerIface)
	}

	return &ResolutionVoteInfo{
		Resolution: Resolution{
			ID:         resolution.ID,
			Type:       resolution.Type,
			Body:       resolution.Body,
			Expiration: resolution.Expiration,
		},
		ApprovedPower: approvedPower,
	}, tx.Commit(ctx)
}

// ContainsBodyOrFinished returns true if (any of the following are true):
// 1. the resolution has a body
// 2. the resolution has expired
// 3. the resolution has been approved
// TODO: should be uncommitted data query
func (v *VoteProcessor) ContainsBodyOrFinished(ctx context.Context, db sql.DB, id types.UUID) (bool, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	// we check for existence of body in resolutions table before checking
	// for the resolution ID in the processed table, since it is a faster lookup.
	// furthermore, we are more likely to hit the resolutions table during consensus,
	// and processed table during catchup. consensus speed is more important.
	res, err := tx.Execute(ctx, getResolutionBody, id[:])
	if err != nil {
		return false, err
	}

	if len(res.Rows) != 0 {
		if len(res.Rows[0]) != 1 {
			// this should never happen, just for safety
			return false, fmt.Errorf("invalid number of columns returned. this is an internal bug")
		}

		if res.Rows[0][0] != nil {
			return true, nil
		}
	}

	processed, err := v.IsProcessed(ctx, tx, id)
	if err != nil {
		return false, err
	}
	return processed, tx.Commit(ctx)
}

// GetVotesByCategory gets all votes of a specific category.
// It does not read uncommitted data. (what is the indented consumer?)
func (v *VoteProcessor) GetVotesByCategory(ctx context.Context, db sql.DB, category string) (votes []*Resolution, err error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Execute(ctx, resolutionsByType, category)
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return []*Resolution{}, nil
	}

	if len(res.Rows[0]) != 4 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	// res.ReturnedColumns[0] == id
	// res.ReturnedColumns[1] == body (can be nil)
	// res.ReturnedColumns[2] == type (can be nil)
	// res.ReturnedColumns[3] == expiration
	votes = make([]*Resolution, len(res.Rows))
	for i, row := range res.Rows {
		resolution, err := getResolutionFromRow(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}

		votes[i] = resolution
	}

	return votes, tx.Commit(ctx)
}

// getResolutionFromRow gets a resolution from a SQL result.
// It assumes that the result has the following columns:
// 0: id
// 1: body
// 2: type
// 3: expiration
func getResolutionFromRow(rows []any) (*Resolution, error) {
	id, ok := rows[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for id (%T)", rows[0])
	}
	body, ok := rows[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for body (%T)", rows[1])
	}
	category, ok := rows[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for category (%T)", rows[2])
	}
	expiration, ok := rows[3].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid type for expiration (%T)", rows[3])
	}

	return &Resolution{
		ID:         types.UUID(id),
		Type:       category,
		Body:       body,
		Expiration: expiration,
	}, nil
}

// UpdateVoter adds a voter to the system, with a given voting power.
// If the voter already exists, it will add the power to the existing power.
// If the power is 0, it will remove the voter from the system.
// If negative power is given, it will subtract the power from the existing power.
func (v *VoteProcessor) UpdateVoter(ctx context.Context, db sql.DB, identifier []byte, power int64) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	uuid := types.NewUUIDV5(identifier)

	if power == 0 {
		_, err = tx.Execute(ctx, removeVoter, uuid[:])
	} else if power < 0 {
		_, err = tx.Execute(ctx, decreaseVoterPower, uuid[:], -power)
	} else {
		_, err = tx.Execute(ctx, upsertVoter, uuid[:], identifier, power)
	}
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetVoterPower gets the power of a voter.
// If the voter does not exist, it will return 0.
func (v *VoteProcessor) GetVoterPower(ctx context.Context, db sql.DB, identifier []byte) (power int64, err error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	uuid := types.NewUUIDV5(identifier)

	res, err := tx.Execute(ctx, getVoterPower, uuid[:])
	if err != nil {
		return 0, err
	}

	if len(res.Columns) != 1 {
		// this should never happen, just for safety
		return 0, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	if len(res.Rows) == 0 {
		return 0, nil
	}

	powerIface := res.Rows[0][0]
	power, ok := sql.Int64(powerIface)
	if !ok {
		return 0, fmt.Errorf("invalid type for power (%T). this is an internal bug", powerIface)
	}

	return power, tx.Commit(ctx)
}

// ProcessConfirmedResolutions processes all resolutions that have exceeded
// the threshold of votes, and have a non-nil body.
// It sorts them lexicographically by ID, and processes them in that order.
func (v *VoteProcessor) ProcessConfirmedResolutions(ctx context.Context, db sql.DB) ([]types.UUID, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	powerRes, err := tx.Execute(ctx, totalPower)
	if err != nil {
		return nil, err
	}

	// NOTE: using the generic interface is a pain since we know the types and
	// should be able to scan directly into a typed variable...

	if len(powerRes.Rows) == 0 {
		return nil, nil // cannot process any resolutions
	}

	if len(powerRes.Rows[0]) != 1 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	powerIface := powerRes.Rows[0][0]       // `numeric` => pgtype.Numeric
	totalPower, ok := sql.Int64(powerIface) // powerIface.(int64)
	if !ok {
		// if it is nil, then no validators have been added yet
		if powerIface == nil {
			return nil, nil
		}

		return nil, fmt.Errorf("invalid type for power needed (%T). this is an internal bug", powerIface)
	}

	// big arithmetic to guarantee accuracy
	bigTotal := big.NewInt(totalPower)
	bigThreshold := big.NewInt(v.percentThreshold)

	scaled := new(big.Int).Mul(bigTotal, bigThreshold)

	// divide by 1000000 to get the threshold
	// this is the amount of power needed to approve a resolution
	result := new(big.Int).Div(scaled, big.NewInt(1000000))
	// intDivUp(2*int64(totalPower), 3)

	res, err := tx.Execute(ctx, getConfirmedResolutions, result.Int64())
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return nil, nil // sp.Commit()
	}

	if len(res.Columns) != 4 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	usedDBIDs := make([]types.UUID, len(res.Rows)) // tracks the uuids of the resolutions we have processed

	for i, row := range res.Rows {
		vote, err := getResolutionFromRow(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}

		payload, ok := registeredPayloads[vote.Type]
		if !ok {
			return nil, fmt.Errorf("payload not registered for type %s", vote.Type)
		}

		err = payload.UnmarshalBinary(vote.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		err = payload.Apply(ctx, tx, Datastores{
			Accounts: v.accounts,
			Engine:   v.engine,
		}, v.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to apply payload: %w", err)
		}

		// id needs to be 16 bytes
		// this should never happen, just for safety
		if len(vote.ID) != 16 { // ???? type UUID [16]byte
			return nil, fmt.Errorf("invalid id length for UUID. required 16 bytes, got %d. this is a bug", len(vote.ID))
		}

		usedDBIDs[i] = vote.ID

		_, err = tx.Execute(ctx, markProcessed, vote.ID[:])
		if err != nil {
			return nil, err
		}
	}

	// delete
	_, err = tx.Execute(ctx, deleteResolutions, types.UUIDArray(usedDBIDs))
	if err != nil {
		return nil, err
	}

	return usedDBIDs, tx.Commit(ctx)
}

// alreadyProcessed checks if a vote has either already succeeded, or expired.
// TODO: need committed or uncommitted data? This method is used in many different contexts.
func (v *VoteProcessor) IsProcessed(ctx context.Context, db sql.DB, resolutionID types.UUID) (bool, error) {
	res, err := db.Execute(ctx, alreadyProcessed, resolutionID[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}

// HasVoted checks if a voter has voted on a resolution.
// NOTE: reads uncommitted/pending data
func (v *VoteProcessor) HasVoted(ctx context.Context, db sql.DB, resolutionID types.UUID, from []byte) (bool, error) {
	userId := types.NewUUIDV5(from)

	res, err := db.Execute(ctx, hasVoted, resolutionID[:], userId[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}
