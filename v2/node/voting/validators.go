package voting

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"kwil/extensions/resolutions"
	"kwil/types"
	"math/big"
)

// this file implements the voting logic for validator approvals

const (
	ValidatorJoinEventType   = "validator_join"
	ValidatorRemoveEventType = "validator_remove"
)

func init() {
	err := resolutions.RegisterResolution(ValidatorJoinEventType, resolutions.ModAdd, resolutions.ResolutionConfig{
		ConfirmationThreshold: big.NewRat(2, 3),
		ResolveFunc: func(ctx context.Context, app *types.App, resolution *resolutions.Resolution, block *types.BlockContext) error {
			joinReq := &UpdatePowerRequest{}
			if err := joinReq.UnmarshalBinary(resolution.Body); err != nil {
				return fmt.Errorf("failed to unmarshal join request: %w", err)
			}

			return app.Validators.SetValidatorPower(ctx, app.DB, joinReq.PubKey, joinReq.Power)
		},
	})
	if err != nil {
		panic(err)
	}

	err = resolutions.RegisterResolution(ValidatorRemoveEventType, resolutions.ModAdd, resolutions.ResolutionConfig{
		ConfirmationThreshold: big.NewRat(2, 3),
		ResolveFunc: func(ctx context.Context, app *types.App, resolution *resolutions.Resolution, block *types.BlockContext) error {
			removeReq := &UpdatePowerRequest{}
			if err := removeReq.UnmarshalBinary(resolution.Body); err != nil {
				return fmt.Errorf("failed to unmarshal remove request: %w", err)
			}
			if removeReq.Power != 0 {
				// this should never happen since UpdatePowerRequest is only used for internal communication
				// between modules. Removes are sent from the client in a separate message.
				return errors.New("remove request with non-zero power")
			}

			return app.Validators.SetValidatorPower(ctx, app.DB, removeReq.PubKey, 0)
		},
	})
	if err != nil {
		panic(err)
	}
}

// UpdatePowerRequest is a request to update a validator's power.
type UpdatePowerRequest struct {
	PubKey []byte
	Power  int64
}

// MarshalBinary returns the binary representation of the join request
// It is deterministic
func (j *UpdatePowerRequest) MarshalBinary() ([]byte, error) {
	powerBts := make([]byte, 8)
	binary.BigEndian.PutUint64(powerBts, uint64(j.Power))
	return append(j.PubKey, powerBts...), nil
}

// UnmarshalBinary unmarshals the join request from its binary representation
func (j *UpdatePowerRequest) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("data too short")
	}
	j.PubKey = data[:len(data)-8]
	j.Power = int64(binary.BigEndian.Uint64(data[len(data)-8:]))
	return nil
}
