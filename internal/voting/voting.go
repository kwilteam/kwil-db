package voting

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
)

// NewVoteProcessor creates a new vote processor.
// It initializes the database with the required tables.
func NewVoteProcessor(ctx context.Context, db Datastore, accounts AccountStore, threshhold int64) (*VoteProcessor, error) {
	sp, err := db.Savepoint()
	if err != nil {
		return nil, err
	}
	defer sp.Rollback()

	_, err = db.Execute(ctx, tableResolutionTypes, nil)
	if err != nil {
		return nil, err
	}

	_, err = db.Execute(ctx, tableResolutions, nil)
	if err != nil {
		return nil, err
	}

	_, err = db.Execute(ctx, tableVotes, nil)
	if err != nil {
		return nil, err
	}

	_, err = db.Execute(ctx, tableVoters, nil)
	if err != nil {
		return nil, err
	}

	_, err = db.Execute(ctx, resolutionTypeIndex, nil)
	if err != nil {
		return nil, err
	}

	_, err = db.Execute(ctx, votesResolutionIndex, nil)
	if err != nil {
		return nil, err
	}

	for name := range registeredPayloads {
		uuid := types.NewUUIDV5([]byte(name))

		_, err = db.Execute(ctx, createResolutionType, map[string]any{
			"$id":   uuid[:],
			"$name": name,
		})
		if err != nil {
			return nil, err
		}
	}

	err = sp.Commit()
	if err != nil {
		return nil, err
	}

	return &VoteProcessor{
		threshhold: threshhold,
		db:         db,
		accounts:   accounts,
	}, nil
}

// VoteProcessor manages in-process votes, and manages how they get processed.
// It is responsible for tracking the relative power of voters, and for
// expiring resolutions.
type VoteProcessor struct {
	// threshhold is the threshhold of votes required to approve a resolution
	// it must be a number between 0 and 1000000.  This defines the percentage
	// of total voting power required to approve a resolution (e.g. 500000 = 50%)
	threshhold int64

	db       Datastore
	accounts AccountStore
}

// ResolutionStatus is a struct that contains information about the status of a resolution
type ResolutionStatus struct {
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
// If the voter has already approved the resolution, it will return an error.
func (v *VoteProcessor) Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error {
	// we need to ensure that the resolution ID exists
	_, err := v.db.Execute(ctx, resolutionIDExists, map[string]interface{}{
		"$id":         resolutionID[:],
		"$expiration": expiration,
	})
	if err != nil {
		return err
	}

	userId := types.NewUUIDV5(from)

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, the following will error
	_, err = v.db.Execute(ctx, addVote, map[string]interface{}{
		"$resolution_id": resolutionID[:],
		"$voter_id":      userId[:],
	})

	// TODO: check for a sql error that indicates that the voter does not exist
	return err
}

// CreateVote creates a vote, by submitting a body of a vote, a topic
// and an expiration.  The expiration should be a blockheight.
func (v *VoteProcessor) CreateVote(ctx context.Context, body []byte, category string, expiration int64) error {
	// ensure that the category exists
	_, ok := registeredPayloads[category]
	if !ok {
		return fmt.Errorf("payload not registered for type %s", category)
	}

	uuid := types.NewUUIDV5(body)
	_, err := v.db.Execute(ctx, upsertResolution, map[string]interface{}{
		"$id":         uuid[:],
		"$body":       body,
		"$type":       category,
		"$expiration": expiration,
	})
	return err
}

// Expire expires all votes at or before a given blockheight.
// All expired votes will be removed from the system.
func (v *VoteProcessor) Expire(ctx context.Context, blockheight int64) error {
	_, err := v.db.Execute(ctx, expireResolutions, map[string]interface{}{
		"$blockheight": blockheight,
	})

	return err
}

// GetResolution gets a resolution, identified by the ID.
// It does not read uncommitted data.
func (v *VoteProcessor) GetResolution(ctx context.Context, id types.UUID) (info *ResolutionStatus, err error) {
	res, err := v.db.Query(ctx, getResolution, map[string]interface{}{
		"$id": id[:],
	})
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		res, err = v.db.Query(ctx, getUnfilledResolution, map[string]interface{}{
			"$id": id[:],
		})
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

		expiration, ok := res.Rows[0][0].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid type for expiration. this is an internal bug")
		}
		approvedPower, ok := res.Rows[0][1].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid type for approved power. this is an internal bug")
		}

		return &ResolutionStatus{
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
		return nil, fmt.Errorf("internal bug: %w", err)
	}
	approvedPower, ok := res.Rows[0][4].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid type for approved power. this is an internal bug")
	}

	return &ResolutionStatus{
		Resolution: Resolution{
			ID:         types.UUID(resolution.ID),
			Type:       resolution.Type,
			Body:       resolution.Body,
			Expiration: resolution.Expiration,
		},
		ApprovedPower: approvedPower,
	}, nil
}

// GetVotesByCategory gets all votes of a specific category.
// It does not read uncommitted data.
func (v *VoteProcessor) GetVotesByCategory(ctx context.Context, category string) (votes []*Resolution, err error) {
	res, err := v.db.Query(ctx, resolutionsByType, map[string]interface{}{
		"$type": category,
	})
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

	return votes, nil
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
		return nil, fmt.Errorf("invalid type for id")
	}
	body, ok := rows[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for body")
	}
	category, ok := rows[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for category")
	}
	expiration, ok := rows[3].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid type for expiration")
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
func (v *VoteProcessor) UpdateVoter(ctx context.Context, identifier []byte, power int64) error {
	uuid := types.NewUUIDV5(identifier)

	if power == 0 {
		_, err := v.db.Execute(ctx, removeVoter, map[string]interface{}{
			"$id": uuid[:],
		})
		return err
	}
	if power < 0 {
		_, err := v.db.Execute(ctx, decreaseVoterPower, map[string]interface{}{
			"$id":    uuid[:],
			"$power": -power,
		})
		return err
	}

	_, err := v.db.Execute(ctx, upsertVoter, map[string]interface{}{
		"$id":    uuid[:],
		"$voter": identifier,
		"$power": power,
	})
	return err
}

// GetVoterPower gets the power of a voter.
// If the voter does not exist, it will return 0.
func (v *VoteProcessor) GetVoterPower(ctx context.Context, identifier []byte) (power int64, err error) {
	uuid := types.NewUUIDV5(identifier)

	res, err := v.db.Query(ctx, getVoterPower, map[string]interface{}{
		"$id": uuid[:],
	})
	if err != nil {
		return 0, err
	}

	if len(res.Columns()) != 1 {
		// this should never happen, just for safety
		return 0, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	if len(res.Rows) == 0 {
		return 0, nil
	}

	power, ok := res.Rows[0][0].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid type for power. this is an internal bug")
	}

	return power, nil
}

// ProcessConfirmedResolutions processes all resolutions that have exceeded
// the threshhold of votes, and have a non-nil body.
// It sorts them lexicographically by ID, and processes them in that order.
func (v *VoteProcessor) ProcessConfirmedResolutions(ctx context.Context) error {
	sp, err := v.db.Savepoint()
	if err != nil {
		return err
	}
	defer sp.Rollback()

	// use execute here to read uncommitted data
	powerRes, err := v.db.Execute(ctx, totalPower, nil)
	if err != nil {
		return err
	}

	if len(powerRes.Rows) == 0 {
		return nil // cannot process any resolutions
	}

	if len(powerRes.Rows[0]) != 1 {
		// this should never happen, just for safety
		return fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	totalPower, ok := powerRes.Rows[0][0].(int64)
	if !ok {
		// if it is nil, then no validators have been added yet
		if powerRes.Rows[0][0] == nil {
			return nil
		}

		return fmt.Errorf("invalid type for power needed. this is an internal bug")
	}

	// big arithmetic to guarantee accuracy
	bigTotal := big.NewInt(totalPower)
	bigThreshold := big.NewInt(v.threshhold)

	scaled := new(big.Int).Mul(bigTotal, bigThreshold)

	// divide by 1000000 to get the threshhold
	// this is the amount of power needed to approve a resolution
	result := new(big.Int).Div(scaled, big.NewInt(1000000))

	res, err := v.db.Execute(ctx, getConfirmedResolutions, map[string]interface{}{
		"$power_needed": result.Int64(),
	})
	if err != nil {
		return err
	}

	if len(res.ReturnedColumns) != 4 {
		// this should never happen, just for safety
		return fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	usedDBIDs := make([][]byte, len(res.Rows)) // tracks the uuids of the resolutions we have processed

	for i, row := range res.Rows {
		vote, err := getResolutionFromRow(row)
		if err != nil {
			return fmt.Errorf("internal bug: %w", err)
		}

		payload, ok := registeredPayloads[vote.Type]
		if !ok {
			return fmt.Errorf("payload not registered for type %s", vote.Type)
		}

		err = payload.UnmarshalBinary(vote.Body)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		err = payload.Apply(ctx, &Datastores{
			Accounts:  v.accounts,
			Databases: v.db,
		})
		if err != nil {
			return fmt.Errorf("failed to apply payload: %w", err)
		}

		// id needs to be 16 bytes
		if len(vote.ID) != 16 {
			return fmt.Errorf("invalid id length for UUID. required 16 bytes, got %d", len(vote.ID))
		}

		usedDBIDs[i] = []byte(vote.ID[:])
	}

	if len(usedDBIDs) == 0 {
		return sp.Commit()
	}
	// delete
	_, err = v.db.Execute(ctx, fmt.Sprintf(deleteResolutions, formatResolutionList(usedDBIDs)), nil)
	if err != nil {
		return err
	}

	return sp.Commit()
}
