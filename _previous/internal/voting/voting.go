package voting

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"slices"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/sql/versioning"
)

const (
	ValidatorVoteBodyBytePrice = 1000      // Per byte cost
	ValidatorVoteIDPrice       = 1000 * 16 // 16 bytes for the UUID
)

// initializeVoteStore initializes the vote store with the required tables. It
// will also create any resolution types that have been registered. NOTE: the
// provided DB is used only for initialization. The store is stateless in the
// application, and this DB connection is not assumed as a dependency.
func initializeVoteStore(ctx context.Context, db sql.TxMaker) error {
	upgradeFns := map[int64]versioning.UpgradeFunc{
		0: initVotingTables,
		1: dropHeight,
		2: dropExtraVoteIDColumn,
	}

	err := versioning.Upgrade(ctx, db, votingSchemaName, upgradeFns, voteStoreVersion)
	if err != nil {
		return fmt.Errorf("failed to initialize or upgrade vote store: %w", err)
	}

	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	compiledResolutions := resolutions.ListResolutions()
	resMap := make(map[string]bool, len(compiledResolutions))
	for _, name := range compiledResolutions {
		resMap[name] = true
	}

	dbResolutions, err := getResolutionTypes(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to get resolution types: %w", err)
	}

	// Ensure that all the resolutions in the database are registered
	for name := range dbResolutions {
		if !resMap[name] {
			return fmt.Errorf("resolution %q is in the database but not registered", name)
		}
	}

	// This is moved out of initVotingTables, to ensure that
	// upgrade logic should just handle the schema changes, but not the data.
	// If the database is initialized with the latest version, but for example,
	// If there is a change in the supported resolution types, the upgrade logic
	// skips the resolution types updates.
	for _, name := range compiledResolutions {
		if dbResolutions[name] {
			continue // already exists
		}
		// add the newly registered resolutions to the database.
		uuid := types.NewUUIDV5([]byte(name))
		fmt.Printf("Creating resolution type %q with UUID %v\n", name, uuid)
		_, err := tx.Execute(ctx, createResolutionType, uuid[:], name)
		if err != nil {
			return err
		}
	}

	if err = ensureInsertResolutionFunc(ctx, tx); err != nil {
		return fmt.Errorf("failed to create insert resolution function: %w", err)
	}

	return tx.Commit(ctx)
}

func ensureInsertResolutionFunc(ctx context.Context, db sql.Executor) error {
	_, err := db.Execute(ctx, sqlInsertResolution)
	return err
}

func getResolutionTypes(ctx context.Context, db sql.Executor) (map[string]bool, error) {
	res, err := db.Execute(ctx, getResolutionTypesQuery)
	if err != nil {
		return nil, err
	}

	var resTypes = make(map[string]bool)
	for _, row := range res.Rows {
		name, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid type for name (%T)", row[0])
		}
		resTypes[name] = true
	}

	return resTypes, nil
}

func initVotingTables(ctx context.Context, db sql.DB) error {
	initStmts := []string{ //createVotingSchema,
		tableVoters, tableResolutionTypes, tableResolutions,
		resolutionsTypeIndex, tableProcessed, tableVotes, tableHeight} // order important

	for _, stmt := range initStmts {
		_, err := db.Execute(ctx, stmt)
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

// ApproveResolution approves a resolution from a voter.
// If the resolution does not yet exist, it will be errored,
// Validators should only vote on existing resolutions.
// If the voter does not exist, an error will be returned.
// If the voter has already approved the resolution, no error will be returned.
// This should not be used if the resolution has already been processed
// (see FilterNotProcessed)
func ApproveResolution(ctx context.Context, db sql.TxMaker, resolutionID *types.UUID, from []byte) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Expectation is that the resolution is already created when the voteBody
	// is submitted. Nodes won't submit the voteIDs for events which don't have
	// resolutions.

	// if the voter does not exist, the following will error
	// if the vote from the voter already exists, nothing will happen
	// if the resolution doesn't exist, the following would error
	userID := types.NewUUIDV5(from)
	_, err = tx.Execute(ctx, addVote, resolutionID[:], userID[:])
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreateResolution creates a resolution for a votable event. The expiration
// should be a block height. Resolution creation will fail if the resolution
// either already exists or has been processed.
func CreateResolution(ctx context.Context, db sql.TxMaker, event *types.VotableEvent, expiration int64, voteBodyProposer []byte) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// NOTE: could check IsProcessed() here and skip the insert.

	id := event.ID()
	_, err = tx.Execute(ctx, insertResolution, id[:], event.Body, event.Type, expiration, voteBodyProposer)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteResolution deletes a resolution from the database by ID if it exists.
func DeleteResolution(ctx context.Context, db sql.TxMaker, id *types.UUID) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Execute(ctx, deleteResolution, id[:])
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
		return nil, fmt.Errorf("invalid length for id, required 16, got (%d)", len(bts))
	}

	uid := types.UUID(slices.Clone(bts))
	v.ID = &uid

	if row[1] == nil {
		v.Body = nil
	} else {
		vBody, ok := row[1].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for body (%T)", row[1])
		}
		v.Body = slices.Clone(vBody)
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

	var voters [][]byte
	voters, ok = row[5].([][]byte)
	if !ok {
		return nil, fmt.Errorf("invalid type for voters (%T)", row[5])
	}

	// returns bigendian int64 + pubKey in []byte
	v.Voters = make([]*types.Validator, 0)
	for _, voter := range voters {
		if voter == nil {
			continue // pgx returns nil aggregates as length one []interface{} with a nil element
		}

		// the first 8 bytes are the power
		if len(voter) < 8 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid length for voter (%d)", len(voter))
		}

		var num uint64
		err := binary.Read(bytes.NewReader(voter[:8]), binary.BigEndian, &num)
		if err != nil {
			return nil, fmt.Errorf("failed to read bigendian int64 from voter: %w", err)
		}

		v.Voters = append(v.Voters, &types.Validator{
			Power:  int64(num),
			PubKey: slices.Clone(voter[8:]),
		})
	}

	if row[6] == nil {
		v.Proposer = nil
	} else {
		vProposer, ok := row[6].([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid type for voteBodyProposer (%T)", row[6])
		}
		v.Proposer = slices.Clone(vProposer)
	}

	return v, nil
}

// GetResolutionInfo gets a resolution, identified by the ID.
func GetResolutionInfo(ctx context.Context, db sql.Executor, id *types.UUID) (*resolutions.Resolution, error) {
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
// less than or equal to the given block height.
func GetExpired(ctx context.Context, db sql.Executor, blockHeight int64) ([]*resolutions.Resolution, error) {
	res, err := db.Execute(ctx, getResolutionsFullInfoByExpiration, blockHeight)
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
	// This is the consensus conn's write txn, so we can't make a RO tx. We
	// create a transaction here in case a query here fails we don't break abci.
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // we can always rollback, since we are only reading

	thresholdPower := RequiredPower(ctx, tx, threshold, totalPower)
	res, err := tx.Execute(ctx, getResolutionsFullInfoByPower, resType, thresholdPower)
	if err != nil {
		return nil, fmt.Errorf("getResolutionsFullInfoByPower: %w", err)
	}

	results := make([]*resolutions.Resolution, len(res.Rows))
	for i, row := range res.Rows {
		results[i], err = fromRow(row)
		if err != nil {
			return nil, fmt.Errorf("fromRow: %w", err)
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

// GetResolutionIDsByTypeAndProposer gets all resolution ids of a specific type and the body proposer.
func GetResolutionIDsByTypeAndProposer(ctx context.Context, db sql.Executor, resType string, proposer []byte) ([]*types.UUID, error) {
	res, err := db.Execute(ctx, getResolutionByTypeAndProposer, resType, proposer)
	if err != nil {
		return nil, err
	}

	ids := make([]*types.UUID, len(res.Rows))

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

		uuid := types.UUID(slices.Clone(id))
		ids[i] = &uuid
	}

	return ids, nil
}

// ResolutionExists checks if a resolution of the given ID exists.
func ResolutionExists(ctx context.Context, db sql.Executor, id *types.UUID) (bool, error) {
	res, err := db.Execute(ctx, resolutionExists, id[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) == 1, nil
}

// DeleteResolutions deletes a slice of resolution IDs from the database.
// It will mark the resolutions as processed in the processed table.
func DeleteResolutions(ctx context.Context, db sql.Executor, ids ...*types.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.Execute(ctx, deleteResolutions, types.UUIDArray(ids).Bytes())
	return err
}

// MarkProcessed marks a set of resolutions as processed.
func MarkProcessed(ctx context.Context, db sql.Executor, ids ...*types.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.Execute(ctx, markManyProcessed, types.UUIDArray(ids).Bytes())
	return err
}

// IsProcessed checks if a vote has been marked as processed.
func IsProcessed(ctx context.Context, tx sql.Executor, resolutionID *types.UUID) (bool, error) {
	res, err := tx.Execute(ctx, alreadyProcessed, resolutionID[:])
	if err != nil {
		return false, err
	}

	return len(res.Rows) != 0, nil
}

// FilterNotProcessed takes a set of resolutions and returns the ones that have not been processed.
// If a resolution does not exist, it WILL be included in the result.
func FilterNotProcessed(ctx context.Context, db sql.Executor, ids []*types.UUID) ([]*types.UUID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	res, err := db.Execute(ctx, returnProcessed, types.UUIDArray(ids).Bytes())
	if err != nil {
		return nil, err
	}

	processed := make(map[types.UUID]bool, len(res.Rows))
	for _, row := range res.Rows {
		if len(row) != 1 {
			// this should never happen, just for safety
			return nil, fmt.Errorf("invalid number of columns returned. this is an internal bug")
		}
		idBts := slices.Clone(row[0].([]byte))
		processed[types.UUID(idBts)] = true
	}

	var notProcessed []*types.UUID
	for _, id := range ids {
		if _, ok := processed[*id]; !ok {
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
			PubKey: slices.Clone(voterBts),
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
func RequiredPower(ctx context.Context, db sql.Executor, threshold *big.Rat, totalPower int64) int64 {
	numerator := threshold.Num().Int64()
	denominator := threshold.Denom().Int64()
	return intDivUpFraction(totalPower, numerator, denominator)
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

func ResolutionStatus(ctx context.Context, db sql.DB, resolution *resolutions.Resolution) (expiresAt int64, board [][]byte, approvals []bool, err error) {
	allVoters, err := GetValidators(ctx, db)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to retrieve voters")
	}

	// to create the board, we will take a list of all approvers and append the voters.
	// we will then remove any duplicates the second time we see them.
	// this will result with all approvers at the start of the list, and all voters at the end.
	// finally, the approvals will be true for the length of the approvers, and false for found.length - voters.length
	board = make([][]byte, 0, len(allVoters))
	approvals = make([]bool, len(allVoters))
	for i, v := range resolution.Voters {
		board = append(board, v.PubKey)
		approvals[i] = true
	}
	for _, v := range allVoters {
		board = append(board, v.PubKey)
	}

	// we will now remove duplicates from the board.
	found := make(map[string]struct{})
	for i := 0; i < len(board); i++ {
		if _, ok := found[string(board[i])]; ok {
			board = append(board[:i], board[i+1:]...)
			i--
			continue
		}
		found[string(board[i])] = struct{}{}
	}

	return resolution.ExpirationHeight, board, approvals, nil
}

func DeleteResolutionsByType(ctx context.Context, db sql.Executor, resTypes []string) error {
	uuids := make([]*types.UUID, len(resTypes))
	for i, resType := range resTypes {
		uuids[i] = types.NewUUIDV5([]byte(resType))
	}
	_, err := db.Execute(ctx, deleteResolutionsByTypeSQL, types.UUIDArray(uuids).Bytes())
	return err
}

func ReadjustExpirations(ctx context.Context, db sql.Executor, startHeight int64) error {
	// Subtracts the start height from the expiration height of all resolutions
	_, err := db.Execute(ctx, readjustExpirationsSQL, startHeight)
	return err
}
