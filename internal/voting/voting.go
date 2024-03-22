package voting

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

const (
	ValidatorVoteBodyBytePrice = 1000      // Per byte cost
	ValidatorVoteIDPrice       = 1000 * 16 // 16 bytes for the UUID
)

// InitializeVoteStore initializes the vote store with the required tables.
// It will also create any resolution types that have been registered.
func InitializeVoteStore(ctx context.Context, db sql.DB) error {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initTables,
		1: dropHeight,
		2: dropExtraVoteIDColumn,
	}

	err := versioning.Upgrade(ctx, db, VotingSchemaName, upgradeFns, voteStoreVersion)
	if err != nil {
		return fmt.Errorf("failed to initialize or upgrade vote store: %w", err)
	}

	return nil
}

func initTables(ctx context.Context, db sql.DB) error {
	initStmts := []string{ //createVotingSchema,
		tableVoters, tableResolutionTypes, tableResolutions,
		resolutionsTypeIndex, tableProcessed, tableVotes, tableHeight} // order important

	for _, stmt := range initStmts {
		_, err := db.Execute(ctx, stmt)
		if err != nil {
			return err
		}
	}

	resolutions := resolutions.ListResolutions()
	for _, name := range resolutions {
		uuid := types.NewUUIDV5([]byte(name))
		_, err := db.Execute(ctx, createResolutionType, uuid[:], name)
		if err != nil {
			return err
		}
	}

	return nil
}

func dropHeight(ctx context.Context, db sql.DB) error {
	_, err := db.Execute(ctx, dropHeightTable)
	return err
}

func dropExtraVoteIDColumn(ctx context.Context, db sql.DB) error {
	_, err := db.Execute(ctx, dropExtraVoteID)
	return err
}

// Approve approves a resolution from a voter.
// If the resolution does not yet exist, it will be errored,
// Validators should only vote on existing resolutions.
// If the voter does not exist, an error will be returned.
// If the voter has already approved the resolution, no error will be returned.
// If the resolution has already been processed, no error will be returned.
func ApproveResolution(ctx context.Context, db sql.TxMaker, resolutionID types.UUID, expiration int64, from []byte) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Expectation is that the resolution is already created when the voteBody is submitted. and nodes wont submit the voteIDs for events which doesn't have resolutions.
	userId := types.NewUUIDV5(from)

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, nothing will happen
	// if the resolution doesn't exist, the following would error
	_, err = tx.Execute(ctx, addVote, resolutionID[:], userId[:])
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreateResolution creates a resolution and subm vote
// The expiration should be a blockheight.
// If the resolution already exists do nothing.
// If the resolution was already processed, nothing will happen.
func CreateResolution(ctx context.Context, db sql.TxMaker, event *types.VotableEvent, expiration int64, voteBodyProposer []byte) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	id := event.ID()

	_, err = tx.Execute(ctx, insertResolution, id[:], event.Body, event.Type, expiration, voteBodyProposer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// fromRow converts a row from the database into a resolutions.Resolution
// It expects there to be 7 columns in the row, in the following order:
// id, body, type, expiration, approved_power, voters, voteBodyProposer
func fromRow(row []any) (*resolutions.Resolution, error) {
	if len(row) != 7 {
		return nil, fmt.Errorf("expected 7 columns, got %d", len(row))
	}

	v := &resolutions.Resolution{}

	bts, ok := row[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for id (%T)", row[0])
	}
	if len(bts) != 16 {
		// this should never happen, just for safety
		return nil, fmt.Errorf("invalid length for id. required 16 bytes, got %d", len(bts))
	}
	v.ID = types.UUID(bts)

	if row[1] == nil {
		v.Body = nil
	} else {
		v.Body, ok = row[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for body (%T)", row[1])
		}
	}

	v.Type, ok = row[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid type for type (%T)", row[2])
	}

	v.ExpirationHeight, ok = sql.Int64(row[3])
	if !ok {
		return nil, fmt.Errorf("invalid type for expiration (%T)", row[3])
	}

	if row[4] == nil {
		v.ApprovedPower = 0
	} else {
		v.ApprovedPower, ok = sql.Int64(row[4])
		if !ok {
			return nil, fmt.Errorf("invalid type for approved_power (%T)", row[4])
		}
	}

	var voters []any
	voters, ok = row[5].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid type for voters (%T)", row[5])
	}

	// returns bigendian int64 + pubKey in []byte
	v.Voters = make([]*types.Validator, 0)
	for _, voter := range voters {
		if voter == nil {
			continue // pgx returns nil aggregates as length one []interface{} with a nil element
		}

		voterBts, ok := voter.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for voter (%T)", voter)
		}

		if len(voterBts) < 8 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid length for voter (%d)", len(voterBts))
		}

		var num uint64
		err := binary.Read(bytes.NewReader(voterBts[:8]), binary.BigEndian, &num)
		if err != nil {
			return nil, fmt.Errorf("failed to read bigendian int64 from voter: %w", err)
		}

		v.Voters = append(v.Voters, &types.Validator{
			Power:  int64(num),
			PubKey: voterBts[8:],
		})
	}

	if row[6] == nil {
		v.Proposer = nil
	} else {
		v.Proposer, ok = row[6].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for voteBodyProposer (%T)", row[6])
		}
	}

	return v, nil
}

// GetResolutionInfo gets a resolution, identified by the ID.
func GetResolutionInfo(ctx context.Context, db sql.Executor, id types.UUID) (*resolutions.Resolution, error) {
	res, err := db.Execute(ctx, getFullResolutionInfo, id[:])
	if err != nil {
		return nil, err
	}

	if len(res.Rows) != 1 {
		return nil, fmt.Errorf("expected 1 row, got %d", len(res.Rows))
	}

	if len(res.Rows[0]) != 7 {
		return nil, fmt.Errorf("expected 7 columns, got %d", len(res.Rows[0]))
	}

	return fromRow(res.Rows[0])
}

// GetExpired returns all resolutions with an expiration
// less than or equal to the given blockheight.
func GetExpired(ctx context.Context, db sql.Executor, blockheight int64) ([]*resolutions.Resolution, error) {
	res, err := db.Execute(ctx, getResolutionsFullInfoByExpiration, blockheight)
	if err != nil {
		return nil, err
	}

	ids := make([]*resolutions.Resolution, len(res.Rows))
	for i, row := range res.Rows {
		ids[i], err = fromRow(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}
	}

	return ids, nil
}

// GetResolutionsByThresholdAndType gets all resolutions that have reached the threshold of votes and are of a specific type.
func GetResolutionsByThresholdAndType(ctx context.Context, db sql.TxMaker, threshold *big.Rat, resType string, totalPower int64) ([]*resolutions.Resolution, error) {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // we can always rollback, since we are only reading

	thresholdPower, err := RequiredPower(ctx, tx, threshold, totalPower)
	if err != nil {
		return nil, err
	}

	res, err := tx.Execute(ctx, getResolutionsFullInfoByPower, resType, thresholdPower)
	if err != nil {
		return nil, err
	}

	results := make([]*resolutions.Resolution, len(res.Rows))
	for i, row := range res.Rows {
		results[i], err = fromRow(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}
	}

	return results, nil
}

// GetResolutionsByType gets all resolutions of a specific type.
func GetResolutionsByType(ctx context.Context, db sql.Executor, resType string) ([]*resolutions.Resolution, error) {
	res, err := db.Execute(ctx, getResolutionsFullInfoByType, resType)
	if err != nil {
		return nil, err
	}

	results := make([]*resolutions.Resolution, len(res.Rows))
	for i, row := range res.Rows {
		results[i], err = fromRow(row)
		if err != nil {
			return nil, fmt.Errorf("internal bug: %w", err)
		}
	}

	return results, nil

}

// DeleteResolutions deletes a slice of resolution IDs from the database.
// It will mark the resolutions as processed in the processed table.
func DeleteResolutions(ctx context.Context, db sql.Executor, ids ...types.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.Execute(ctx, deleteResolutions, types.UUIDArray(ids))
	return err
}

// MarkProcessed marks a set of resolutions as processed.
func MarkProcessed(ctx context.Context, db sql.Executor, ids ...types.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.Execute(ctx, markManyProcessed, types.UUIDArray(ids))
	return err
}

// IsProcessed checks if a vote has been marked as processed.
func IsProcessed(ctx context.Context, tx sql.Executor, resolutionID types.UUID) (bool, error) {
	res, err := tx.Execute(ctx, alreadyProcessed, resolutionID[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}

// FilterNotProcessed takes a set of resolutions and returns the ones that have not been processed.
// If a resolution does not exist, it WILL be included in the result.
func FilterNotProcessed(ctx context.Context, db sql.Executor, ids []types.UUID) ([]types.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	res, err := db.Execute(ctx, returnProcessed, types.UUIDArray(ids))
	if err != nil {
		return nil, err
	}

	processed := make(map[types.UUID]bool, len(res.Rows))
	for _, row := range res.Rows {
		if len(row) != 1 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
		}
		processed[types.UUID(row[0].([]byte))] = true
	}

	var notProcessed []types.UUID
	for _, id := range ids {
		if _, ok := processed[id]; !ok {
			notProcessed = append(notProcessed, id)
		}
	}
	return notProcessed, nil
}

// GetValidatorPower gets the power of a voter.
// If the voter does not exist, it will return 0.
func GetValidatorPower(ctx context.Context, db sql.Executor, identifier []byte) (power int64, err error) {
	uuid := types.NewUUIDV5(identifier)

	res, err := db.Execute(ctx, getVoterPower, uuid[:])
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

// GetValidators gets all voters in the vote store, along with their power.
func GetValidators(ctx context.Context, db sql.Executor) ([]*types.Validator, error) {
	res, err := db.Execute(ctx, allVoters)
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

	voters := make([]*types.Validator, len(res.Rows))
	for i, row := range res.Rows {
		if len(row) != 2 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
		}

		voterBts, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for voter")
		}
		power, ok := sql.Int64(row[1])
		if !ok {
			return nil, fmt.Errorf("invalid type for power")
		}
		voters[i] = &types.Validator{
			PubKey: voterBts,
			Power:  power,
		}
	}

	return voters, nil
}

// SetValidatorPower sets the power of a voter.
// It will create the voter if it does not exist.
// It will return an error if a negative power is given.
// If set to 0, the voter will be deleted.
func SetValidatorPower(ctx context.Context, db sql.Executor, recipient []byte, power int64) error {
	if power < 0 {
		return fmt.Errorf("cannot set a negative power")
	}

	uuid := types.NewUUIDV5(recipient)

	if power == 0 {
		_, err := db.Execute(ctx, removeVoter, uuid[:])
		return err
	}

	_, err := db.Execute(ctx, upsertVoter, uuid[:], recipient, power)
	return err
}

// RequiredPower gets the required power to meet the threshold requirements.
func RequiredPower(ctx context.Context, db sql.Executor, threshold *big.Rat, totalPower int64) (int64, error) {
	numerator := threshold.Num().Int64()
	denominator := threshold.Denom().Int64()

	result := intDivUpFraction(totalPower, numerator, denominator)
	return result, nil
}

// GetResolutionIDsByTypeAndProposer gets all resolution ids of a specific type and the body proposer.
func GetResolutionIDsByTypeAndProposer(ctx context.Context, db sql.Executor, resType string, proposer []byte) ([]types.UUID, error) {
	res, err := db.Execute(ctx, getResolutionByTypeAndProposer, resType, proposer)
	if err != nil {
		return nil, err
	}

	ids := make([]types.UUID, len(res.Rows))

	if len(res.Rows) == 0 {
		return ids, nil
	}

	for i, row := range res.Rows {
		id, ok := row[0].([]byte)
		if !ok {
			return nil, fmt.Errorf("internal bug: invalid type for id (%T)", row[0])
		}
		if len(id) != 16 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("internal bug: invalid length for id. required 16 bytes, got %d", len(id))
		}

		ids[i] = types.UUID(id)
	}

	return ids, nil
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
