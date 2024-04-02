package statesync

import (
	"errors"
	"fmt"

	abciTypes "github.com/cometbft/cometbft/abci/types"
)

var (
	ErrStateSyncInProgress    = fmt.Errorf("statesync already in progress")
	ErrStateSyncNotInProgress = fmt.Errorf("statesync not in progress")

	ErrAbortSnapshot             = fmt.Errorf("abort snapshot")
	ErrRejectSnapshot            = fmt.Errorf("reject snapshot")
	ErrUnsupportedSnapshotFormat = fmt.Errorf("unsupported snapshot format")
	ErrInvalidSnapshot           = fmt.Errorf("invalid snapshot")

	ErrAbortSnapshotChunk  = fmt.Errorf("abort snapshot chunk")
	ErrRetrySnapshotChunk  = fmt.Errorf("retry snapshot chunk") // retries without refetching the chunk
	ErrRetrySnapshot       = fmt.Errorf("retry snapshot")
	ErrRejectSnapshotChunk = fmt.Errorf("reject snapshot chunk")
)

func ToABCIOfferSnapshotResponse(err error) abciTypes.ResponseOfferSnapshot_Result {

	if errors.Is(err, nil) {
		return abciTypes.ResponseOfferSnapshot_ACCEPT
	} else if errors.Is(err, ErrAbortSnapshot) {
		return abciTypes.ResponseOfferSnapshot_ABORT
	} else if errors.Is(err, ErrRejectSnapshot) {
		return abciTypes.ResponseOfferSnapshot_REJECT
	} else if errors.Is(err, ErrUnsupportedSnapshotFormat) {
		return abciTypes.ResponseOfferSnapshot_REJECT_FORMAT
	} else {
		return abciTypes.ResponseOfferSnapshot_UNKNOWN
	}
}

func ToABCIApplySnapshotChunkResponse(err error) abciTypes.ResponseApplySnapshotChunk_Result {

	if errors.Is(err, nil) {
		return abciTypes.ResponseApplySnapshotChunk_ACCEPT
	} else if errors.Is(err, ErrAbortSnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_ABORT
	} else if errors.Is(err, ErrRetrySnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	} else if errors.Is(err, ErrRejectSnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
	} else {
		return abciTypes.ResponseApplySnapshotChunk_UNKNOWN
	}
}
