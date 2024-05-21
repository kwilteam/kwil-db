package statesync

import (
	"errors"
	"fmt"

	cmtAPIabci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
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

func ToABCIOfferSnapshotResult(err error) cmtAPIabci.OfferSnapshotResult {
	if err == nil {
		return cmtAPIabci.OFFER_SNAPSHOT_RESULT_ACCEPT
	}

	if errors.Is(err, ErrAbortSnapshot) {
		return cmtAPIabci.OFFER_SNAPSHOT_RESULT_ABORT
	}

	if errors.Is(err, ErrRejectSnapshot) {
		return cmtAPIabci.OFFER_SNAPSHOT_RESULT_REJECT
	}

	if errors.Is(err, ErrUnsupportedSnapshotFormat) {
		return cmtAPIabci.OFFER_SNAPSHOT_RESULT_REJECT_FORMAT
	}

	return cmtAPIabci.OFFER_SNAPSHOT_RESULT_UNKNOWN
}

func ToABCIApplySnapshotChunkResponse(err error) cmtAPIabci.ApplySnapshotChunkResult {

	if err == nil {
		return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_ACCEPT
	}

	if errors.Is(err, ErrAbortSnapshotChunk) {
		return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_ABORT
	}

	if errors.Is(err, ErrRetrySnapshotChunk) {
		return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY
	}

	if errors.Is(err, ErrRefetchSnapshotChunk) {
		return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY
	}

	if errors.Is(err, ErrRejectSnapshotChunk) {
		return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT
	}

	// if errors.Is(err, ErrRetrySnapshot) {
	// 	return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_RETRY_SNAPSHOT
	// }

	// If the error is unrecognized, fall back to rejecting the snapshot.
	// Returning UNKNOWN is fatal to cometbft.
	return cmtAPIabci.APPLY_SNAPSHOT_CHUNK_RESULT_REJECT_SNAPSHOT
}
