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

	ErrAbortSnapshotChunk   = fmt.Errorf("abort snapshot chunk")
	ErrRetrySnapshotChunk   = fmt.Errorf("retry snapshot chunk")   // retries without refetching the chunk
	ErrRefetchSnapshotChunk = fmt.Errorf("refetch snapshot chunk") // retries after refetching the chunk
	ErrRejectSnapshotChunk  = fmt.Errorf("reject snapshot chunk")
	// ErrRetrySnapshot        = fmt.Errorf("retry snapshot") // request full retry of snapshot ==> ResponseApplySnapshotChunk_RETRY_SNAPSHOT
)

func ToABCIOfferSnapshotResponse(err error) abciTypes.ResponseOfferSnapshot_Result {
	if err == nil {
		return abciTypes.ResponseOfferSnapshot_ACCEPT
	}

	if errors.Is(err, ErrAbortSnapshot) {
		return abciTypes.ResponseOfferSnapshot_ABORT
	}

	if errors.Is(err, ErrRejectSnapshot) {
		return abciTypes.ResponseOfferSnapshot_REJECT
	}

	if errors.Is(err, ErrUnsupportedSnapshotFormat) {
		return abciTypes.ResponseOfferSnapshot_REJECT_FORMAT
	}

	return abciTypes.ResponseOfferSnapshot_UNKNOWN
}

func ToABCIApplySnapshotChunkResponse(err error) abciTypes.ResponseApplySnapshotChunk_Result {

	if err == nil {
		return abciTypes.ResponseApplySnapshotChunk_ACCEPT
	}

	if errors.Is(err, ErrAbortSnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_ABORT
	}

	if errors.Is(err, ErrRetrySnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	}

	if errors.Is(err, ErrRefetchSnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_RETRY
	}

	if errors.Is(err, ErrRejectSnapshotChunk) {
		return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
	}

	// if errors.Is(err, ErrRetrySnapshot) {
	// 	return abciTypes.ResponseApplySnapshotChunk_RETRY_SNAPSHOT
	// }

	// If the error is unrecognized, fall back to rejecting the snapshot.
	// Returning UNKNOWN is fatal to cometbft.
	return abciTypes.ResponseApplySnapshotChunk_REJECT_SNAPSHOT
}
