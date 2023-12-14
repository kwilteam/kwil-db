package voting

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
)

// VoteProcessor manages in-process votes, and manages how they get processed.
// It is responsible for tracking the relative power of voters, and for
// expiring resolutions.
type VoteProcessor struct {
	// threshhold is the threshhold of votes required to approve a resolution
	// it must be a number between 0 and 1000000.  This defines the percentage
	// of total voting power required to approve a resolution (e.g. 500000 = 50%)
	threshhold int64

	db Datastore
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

type Voter struct {
	// Name is the name of the voter
	// This is either a public key or address
	Name string

	// Power is the voting power of the voter
	Power int64
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
		"$id":         resolutionID,
		"$expiration": expiration,
	})
	if err != nil {
		return err
	}

	userId := types.NewUUIDV5(from)

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, the following will error
	_, err = v.db.Execute(ctx, addVote, map[string]interface{}{
		"$resolution_id": resolutionID,
		"$voter_id":      userId,
	})

	// TODO: check for a sql error that indicates that the voter does not exist
	return err
}

// CreateVote creates a vote, by submitting a body of a vote, a topic
// and an expiration.  The expiration should be a blockheight.
func (v *VoteProcessor) CreateVote(ctx context.Context, body []byte, category string, expiration int64) error {
	_, err := v.db.Execute(ctx, upsertResolution, map[string]interface{}{
		"$id":         types.NewUUIDV5(body),
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
func (v *VoteProcessor) GetResolution(ctx context.Context, id types.UUID) (info *ResolutionStatus, err error) {
	res, err := v.db.Query(ctx, getResolution, map[string]interface{}{
		"$id": id,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return nil, ErrResolutionNotFound
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

	resolution, err := getResolutionFromRows(res.Rows[0])
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

// GetVotesByCategory gets all votes of a specific category
func (v *VoteProcessor) GetVotesByCategory(ctx context.Context, category string) (votes []*Resolution, err error) {
	res, err := v.db.Query(ctx, resolutionsByType, map[string]interface{}{
		"$category": category,
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
		resolution, err := getResolutionFromRows(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}

		votes[i] = resolution
	}

	return votes, nil
}

// getResolutionFromRows gets a resolution from a SQL result.
// It assumes that the result has the following columns:
// 0: id
// 1: body
// 2: type
// 3: expiration
func getResolutionFromRows(rows []any) (*Resolution, error) {
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

// AddVoter adds a voter to the system, with a given voting power.
// If the voter already exists, it will update the voting power.
func (v *VoteProcessor) AddVoter(ctx context.Context, identifier []byte, power int64) error {
	_, err := v.db.Execute(ctx, upsertVoter, map[string]interface{}{
		"$id":    types.NewUUIDV5(identifier),
		"$name":  identifier,
		"$power": power,
	})
	return err
}

// RemoveVoter removes a voter from the system.
// If the voter does not exist, it does nothing.
func (v *VoteProcessor) RemoveVoter(ctx context.Context, identifier []byte) error {
	_, err := v.db.Execute(ctx, removeVoter, map[string]interface{}{
		"$id": types.NewUUIDV5(identifier),
	})
	return err
}

// TODO: we need a way to process confirmed votes
