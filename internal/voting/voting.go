package voting

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

const (
	ValidatorVoteBodyBytePrice = 1000      // Per byte cost
	ValidatorVoteIDPrice       = 1000 * 16 // 16 bytes for the UUID
)

// VoteStore is a connection to a database with read-write access.
type VoteStore interface {
	Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error)
	Query(ctx context.Context, query string, args ...any) (*sql.ResultSet, error)
}

// Threshold is a struct that contains the numerator and denominator of a fraction
// It is used to define the percentage of total voting power required for taking a specific action
type Threshold struct {
	Num   int64
	Denom int64
}

// NewVoteProcessor creates a new vote processor.
// It initializes the database with the required tables.
// The threshold is the percentThreshold of votes required to approve a resolution
// It must be an integer between 0 and 1000000.  This defines the percentage
func NewVoteProcessor(ctx context.Context, db VoteStore, accounts AccountStore,
	reg Datasets, threshold Threshold, logger log.Logger) (*VoteProcessor, error) {
	initStmts := []string{createVotingSchema, tableResolutionTypes, tableVoters, tableResolutions, resolutionsTypeIndex, tableProcessed, tableVotes} // order important
	for _, stmt := range initStmts {
		_, err := db.Execute(ctx, stmt)
		if err != nil {
			return nil, err
		}
	}

	for name := range registeredPayloads {
		uuid := types.NewUUIDV5([]byte(name))
		_, err := db.Execute(ctx, createResolutionType, uuid[:], name)
		if err != nil {
			return nil, err
		}
	}

	return &VoteProcessor{
		percentThreshold: threshold,
		expiryRefundThreshold: Threshold{
			Num:   1,
			Denom: 3,
		}, // 33.3333%
		db:       db,
		accounts: accounts,
		registry: reg,
		logger:   logger,
	}, nil
}

// VoteProcessor manages in-process votes, and manages how they get processed.
// It is responsible for tracking the relative power of voters, and for
// expiring resolutions.
type VoteProcessor struct {
	// percentThreshold is the percentThreshold of votes required to approve a resolution
	// it must be a number between 0 and 1000000.  This defines the percentage
	// of total voting power required to approve a resolution (e.g. 500000 = 50%)
	percentThreshold Threshold

	// expiryRefundThreshold is the percentThreshold of votes required to refund the voters upon expiry of a resolution
	expiryRefundThreshold Threshold

	db       VoteStore
	accounts AccountStore
	registry Datasets // dataset registry with dbid args on methods

	logger log.Logger
}

// ResolutionVoteInfo is a struct that contains information about the status of a resolution
type ResolutionVoteInfo struct {
	Resolution

	// ApprovedPower is the aggregate amount of power that has approved
	// the vote.
	ApprovedPower int64

	// Proposer
	VoteBodyProposer []byte

	// Voters
	Voters []Voter

	// used to indicate if proposer has submitted both the Body and ID transactions
	SubmittedBodyAndID bool
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
func (v *VoteProcessor) Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error {
	alreadyProcessed, err := v.IsProcessed(ctx, resolutionID)
	if err != nil {
		return err
	}
	if alreadyProcessed {
		return nil
	}

	// we need to ensure that the resolution ID exists
	_, err = v.db.Execute(ctx, resolutionIDExists, resolutionID[:], expiration)
	if err != nil {
		return err
	}

	userId := types.NewUUIDV5(from)

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, the following will error
	_, err = v.db.Execute(ctx, addVote, resolutionID[:], userId[:])
	return err
}

// CreateResolution creates a vote, by submitting a body of a vote, a topic
// and an expiration.  The expiration should be a blockheight.
// If the resolution already exists, it will not be changed.
// If the resolution was already processed, nothing will happen.
func (v *VoteProcessor) CreateResolution(ctx context.Context, event *types.VotableEvent, expiration int64, voteBodyProposer []byte) error {
	alreadyProcessed, err := v.IsProcessed(ctx, event.ID())
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

	proposerId := types.NewUUIDV5(voteBodyProposer)
	// ensure that the proposer exists
	_, err = v.db.Query(ctx, getVoterPower, proposerId[:])
	if err != nil {
		return fmt.Errorf("proposer does not exist: %w", err)
	}

	id := event.ID()

	// Check if the proposer has already submitted the VoteID transaction
	// if yes, update the extraVoteID in the resolutions table so that the node can be refunded correctly.
	voted, err := v.HasVoted(ctx, id, voteBodyProposer)
	if err != nil {
		return err
	}

	_, err = v.db.Execute(ctx, upsertResolution, id[:], event.Body, event.Type, expiration, proposerId[:], voted)
	return err
}

// Expire expires all votes at or before a given blockheight.
// All expired votes will be removed from the system.
// If a resolution has at least 1/3rd of the total voting power, the voters
// will be refunded back the Tx costs.
func (v *VoteProcessor) Expire(ctx context.Context, blockHeight int64) error {
	// Get the total voting power
	totalVotingPower, err := v.RequiredPower(ctx, v.expiryRefundThreshold.Num, v.expiryRefundThreshold.Denom)
	if err != nil {
		return err
	}

	if totalVotingPower == 0 {
		return nil
	}

	// get all expired resolutions
	res, err := v.db.Query(ctx, expiredResolutions, blockHeight)
	if err != nil {
		return err
	}

	if len(res.Rows) == 0 {
		return nil
	}

	if len(res.Rows[0]) != 7 {
		// this should never happen, just for safety
		return fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	ids := make([]types.UUID, len(res.Rows))
	for i, row := range res.Rows {
		// Refund the voters the Tx costs if they have atleast 1/3rd of the total voting power
		resolution, err := v.getResolutionVoteInfoFromRow(ctx, row)
		if err != nil {
			return fmt.Errorf("internal bug: %w", err)
		}
		ids[i] = resolution.ID

		// If the approved power is atleast 1/3rd of the total voting power, refund the voters
		if resolution.ApprovedPower >= totalVotingPower {
			// Refund the voters
			ProposerFee := big.NewInt(int64(len(resolution.Body)) * ValidatorVoteBodyBytePrice)
			VoterFee := big.NewInt(ValidatorVoteIDPrice)

			for _, voter := range resolution.Voters {
				// check if the voter is the proposer
				if bytes.Equal(voter.PubKey, resolution.VoteBodyProposer) {
					credit := ProposerFee
					if resolution.SubmittedBodyAndID {
						credit = credit.Add(credit, VoterFee)
					}
					err = v.accounts.Credit(ctx, voter.PubKey, credit)
					if err != nil {
						return err
					}
				} else {
					err = v.accounts.Credit(ctx, voter.PubKey, VoterFee)
					if err != nil {
						return err
					}
				}
			}
		}

		// Mark the resolution as processed
		_, err = v.db.Execute(ctx, markProcessed, resolution.ID[:])
		if err != nil {
			return err
		}
	}

	// Delete all the expired resolutions
	_, err = v.db.Execute(ctx, deleteResolutions, types.UUIDArray(ids))
	if err != nil {
		return err
	}

	return nil // sp.Commit()
}

// ProcessConfirmedResolutions processes all resolutions that have exceeded
// the threshold of votes, and have a non-nil body.
// It sorts them lexicographically by ID, and processes them in that order.
func (v *VoteProcessor) ProcessConfirmedResolutions(ctx context.Context) ([]types.UUID, error) {
	thresholdPower, err := v.RequiredPower(ctx, v.percentThreshold.Num, v.percentThreshold.Denom)
	if err != nil {
		return nil, err
	}
	if thresholdPower == 0 {
		return nil, nil
	}

	res, err := v.db.Execute(ctx, getConfirmedResolutions, thresholdPower)
	if err != nil {
		return nil, err
	}

	if len(res.Columns) != 7 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned while querying for confirmed resolutions. this is an internal bug")
	}

	usedDBIDs := make([]types.UUID, len(res.Rows)) // tracks the uuids of the resolutions we have processed

	for i, row := range res.Rows {
		voteInfo, err := v.getResolutionVoteInfoFromRow(ctx, row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}

		vote := voteInfo.Resolution

		payload, ok := registeredPayloads[vote.Type]
		if !ok {
			return nil, fmt.Errorf("payload not registered for type %s", vote.Type)
		}

		err = payload.UnmarshalBinary(vote.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		Proposerfee := big.NewInt(int64(len(vote.Body)) * ValidatorVoteBodyBytePrice)
		VoterFee := big.NewInt(ValidatorVoteIDPrice)

		// Refund Voters the Tx cost.
		for _, voter := range voteInfo.Voters {
			// mint()
			if bytes.Equal(voter.PubKey, voteInfo.VoteBodyProposer) {
				refund := Proposerfee
				if voteInfo.SubmittedBodyAndID {
					// If the proposer has submitted both the body and ID transactions, refund the proposer the ID fee as well
					refund = new(big.Int).Add(refund, VoterFee)
				}
				err = v.accounts.Credit(ctx, voter.PubKey, refund)
				if err != nil {
					return nil, fmt.Errorf("failed to credit proposer: %w", err)
				}
			} else {
				err = v.accounts.Credit(ctx, voter.PubKey, VoterFee)
				if err != nil {
					return nil, fmt.Errorf("failed to credit voter: %w", err)
				}
			}
		}

		// 	 reward the proposer and voters.
		err = payload.Apply(ctx, Datastores{
			Accounts:  v.accounts,
			Databases: v.registry,
		}, voteInfo.VoteBodyProposer, voteInfo.Voters, v.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to apply payload: %w", err)
		}

		// id needs to be 16 bytes
		// this should never happen, just for safety
		if len(vote.ID) != 16 { // ???? type UUID [16]byte
			return nil, fmt.Errorf("invalid id length for UUID. required 16 bytes, got %d. this is a bug", len(vote.ID))
		}

		usedDBIDs[i] = vote.ID

		_, err = v.db.Execute(ctx, markProcessed, vote.ID[:])
		if err != nil {
			return nil, err
		}
	}

	if len(usedDBIDs) == 0 {
		return nil, nil //sp.Commit()
	}
	// delete
	_, err = v.db.Execute(ctx, deleteResolutions, types.UUIDArray(usedDBIDs))
	if err != nil {
		return nil, err
	}

	return usedDBIDs, nil // sp.Commit()
}

// ContainsBody returns true if the resolution has a body.
func (v *VoteProcessor) ContainsBody(ctx context.Context, id types.UUID) (bool, error) {
	res, err := v.db.Query(ctx, containsBody, id[:])
	if err != nil {
		return false, err
	}

	if len(res.Rows) == 0 {
		return false, nil
	}

	if len(res.Rows[0]) != 1 {
		// this should never happen, just for safety
		return false, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}
	containsBody, ok := res.Rows[0][0].(bool)
	if !ok {
		return false, fmt.Errorf("invalid type for containsBody (%T). this is an internal bug", res.Rows[0][0])
	}

	return containsBody, nil
}

// ContainsBodyOrFinished returns true if (any of the following are true):
// 1. the resolution has a body
// 2. the resolution has expired
// 3. the resolution has been approved
func (v *VoteProcessor) ContainsBodyOrFinished(ctx context.Context, id types.UUID) (bool, error) {
	// we check for existence of body in resolutions table before checking
	// for the resolution ID in the processed table, since it is a faster lookup.
	// furthermore, we are more likely to hit the resolutions table during consensus,
	// and processed table during catchup. consensus speed is more important.
	containsBody, err := v.ContainsBody(ctx, id)
	if err != nil {
		return false, err
	}

	if containsBody {
		return true, nil
	}

	processed, err := v.IsProcessed(ctx, id)
	if err != nil {
		return false, err
	}

	if processed {
		return true, nil
	}

	return false, nil
}

// alreadyProcessed checks if a vote has either already succeeded, or expired.
func (v *VoteProcessor) IsProcessed(ctx context.Context, resolutionID types.UUID) (bool, error) {
	res, err := v.db.Query(ctx, alreadyProcessed, resolutionID[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}

// HasVoted checks if a voter has voted on a resolution.
func (v *VoteProcessor) HasVoted(ctx context.Context, resolutionID types.UUID, from []byte) (bool, error) {
	userId := types.NewUUIDV5(from)

	res, err := v.db.Query(ctx, hasVoted, resolutionID[:], userId[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}

// GetResolutionVoteInfo gets a resolution, identified by the ID.
// It does not read uncommitted data.
func (v *VoteProcessor) GetResolutionVoteInfo(ctx context.Context, id types.UUID) (info *ResolutionVoteInfo, err error) {
	res, err := v.db.Query(ctx, getResolutionVoteInfo, id[:]) // TODO: register our UUID with scanner's type map?
	if err != nil {
		return nil, err
	}

	// if no rows, then the resolution may still exist, but does not have a body
	if len(res.Rows) == 0 {
		res, err = v.db.Query(ctx, getUnfilledResolutionVoteInfo, id[:])
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

		voters, err := v.getVoters(ctx, id)
		if err != nil {
			return nil, err
		}

		return &ResolutionVoteInfo{
			Resolution: Resolution{
				ID:         id,
				Expiration: expiration,
			},
			ApprovedPower:      approvedPower,
			VoteBodyProposer:   nil,
			Voters:             voters,
			SubmittedBodyAndID: false,
		}, nil
	}

	return v.getResolutionVoteInfoFromRow(ctx, res.Rows[0])
}

// GetVotesByCategory gets all votes of a specific category.
// It does not read uncommitted data. (what is the indented consumer?)
func (v *VoteProcessor) GetVotesByCategory(ctx context.Context, category string) (votes []*Resolution, err error) {
	res, err := v.db.Query(ctx, resolutionsByType, category)
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
	resolution := &Resolution{}

	id, ok := rows[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for id (%T)", rows[0])
	}
	resolution.ID = types.UUID(id)

	if rows[1] != nil {
		body, ok := rows[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for body(%T)", rows[1])
		}
		resolution.Body = body
	}

	if rows[2] != nil {
		category, ok := rows[2].(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for category (%T)", rows[2])
		}
		resolution.Type = category
	}

	expiration, ok := rows[3].(int64)
	if !ok {
		return nil, fmt.Errorf("invalid type for expiration (%T)", rows[3])
	}
	resolution.Expiration = expiration

	return resolution, nil
}

func (v *VoteProcessor) getResolutionVoteInfoFromRow(ctx context.Context, rows []any) (*ResolutionVoteInfo, error) {
	vote := &ResolutionVoteInfo{}

	// ReturnedColumns[0] == id
	// ReturnedColumns[1] == body (can be nil)
	// ReturnedColumns[2] == type (can be nil)
	// ReturnedColumns[3] == expiration
	// ReturnedColumns[4] == approved_power
	// ReturnedColumns[5] == voteBodyProposer
	// ReturnedColumns[6] == extra_vote_id
	if len(rows) != 7 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	res, err := getResolutionFromRow(rows)
	if err != nil {
		return nil, err
	}
	vote.Resolution = *res

	// ApprovedPower
	if rows[4] != nil {
		approvedPower, ok := sql.Int64(rows[4])
		if !ok {
			return nil, fmt.Errorf("invalid type for approved power (%T)", rows[4])
		}
		vote.ApprovedPower = approvedPower
	}

	// voteBodyProposer
	if rows[5] != nil {
		voteBodyProposer, ok := rows[5].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for proposer (%T)", rows[5])
		}
		// Get proposer address
		result, err := v.db.Query(ctx, getVoterName, voteBodyProposer)
		if err != nil {
			return nil, err
		}

		if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
			return nil, fmt.Errorf("voteBodyProposer does not exist")
		}

		proposerAddress, ok := result.Rows[0][0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for proposer address (%T)", result.Rows[0][0])
		}
		vote.VoteBodyProposer = proposerAddress
	}

	// submittedBodyAndID
	if rows[6] != nil {
		submittedBodyAndID, ok := rows[6].(bool)
		if !ok {
			return nil, fmt.Errorf("invalid type for submittedBodyAndID (%T)", rows[6])
		}
		vote.SubmittedBodyAndID = submittedBodyAndID
	}

	voters, err := v.getVoters(ctx, vote.ID)
	if err != nil {
		return nil, err
	}
	vote.Voters = voters

	return vote, nil
}

func (v *VoteProcessor) getVoters(ctx context.Context, resolutionID types.UUID) ([]Voter, error) {
	// Extract Voters along with their power
	res, err := v.db.Execute(ctx, getResolutionVoters, resolutionID[:])
	if err != nil {
		return nil, err
	}

	if len(res.Rows) == 0 {
		return nil, nil
	}

	if len(res.Rows[0]) != 2 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
	}

	voters := make([]Voter, len(res.Rows))
	for i, row := range res.Rows {
		voterBts, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for voter")
		}
		power, ok := sql.Int64(row[1])
		if !ok {
			return nil, fmt.Errorf("invalid type for power")
		}
		voters[i] = Voter{
			PubKey: voterBts,
			Power:  power,
		}
	}

	return voters, nil
}

// UpdateVoter adds a voter to the system, with a given voting power.
// If the voter already exists, it will add the power to the existing power.
// If the power less than or equal to 0, it will remove the voter from the system.
func (v *VoteProcessor) UpdateVoter(ctx context.Context, identifier []byte, power int64) error {
	uuid := types.NewUUIDV5(identifier)

	if power <= 0 {
		_, err := v.db.Execute(ctx, removeVoter, uuid[:])
		return err
	}

	_, err := v.db.Execute(ctx, upsertVoter, uuid[:], identifier, power)
	return err
}

// GetVoterPower gets the power of a voter.
// If the voter does not exist, it will return 0.
// TODO: committed or uncommitted? who's the consumer?
func (v *VoteProcessor) GetVoterPower(ctx context.Context, identifier []byte) (power int64, err error) {
	uuid := types.NewUUIDV5(identifier)

	res, err := v.db.Query(ctx, getVoterPower, uuid[:])
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

	return power, nil
}

// requiredPower gets the required power to meet the threshold requirements.
func (v *VoteProcessor) RequiredPower(ctx context.Context, num, denom int64) (int64, error) {
	powerRes, err := v.db.Execute(ctx, totalPower)
	if err != nil {
		return 0, err
	}

	if len(powerRes.Rows) == 0 {
		return 0, nil // cannot process any resolutions
	}

	if len(powerRes.Rows[0]) != 1 {
		// this should never happen, just for safety
		return 0, fmt.Errorf("invalid number of columns returned while querying total Power. this is an internal bug")
	}

	powerIface := powerRes.Rows[0][0]       // `numeric` => pgtype.Numeric
	totalPower, ok := sql.Int64(powerIface) // powerIface.(int64)
	if !ok {
		// if it is nil, then no validators have been added yet
		if powerRes.Rows[0][0] == nil {
			return 0, nil
		}

		return 0, fmt.Errorf("invalid type for power needed (%T)", powerIface)
	}
	result := intDivUpFraction(totalPower, num, denom)
	return result, nil
}

// intDivUpFraction performs an integer division of (numerator * multiplier /
// divisor) that rounds up. This function is useful for scaling a value by a
// fraction without losing precision due to integer division.
func intDivUpFraction(val, numerator, divisor int64) int64 {
	valBig, numerBig, divBig := big.NewInt(val), big.NewInt(numerator), big.NewInt(divisor)
	// (numerator * val + divisor - 1) / divisor
	tempNumerator := new(big.Int).Mul(numerBig, valBig)
	tempNumerator.Add(tempNumerator, new(big.Int).Sub(divBig, big.NewInt(1)))
	return new(big.Int).Div(tempNumerator, divBig).Int64()
}

// DropAllTables is used to drop all the known tables in the voting store
// postgresql scheme.
func DropAllTables(ctx context.Context, db VoteStore) error {
	const dropTableTmpl = `DROP TABLE IF EXISTS ` + votingSchemaName + `.%s CASCADE;`
	for _, tableName := range []string{"resolutions", "resolution_types", "voters", "votes", "processed"} {
		_, err := db.Execute(ctx, fmt.Sprintf(dropTableTmpl, tableName))
		if err != nil {
			return err
		}
	}
	return nil
}
