package voting

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/core/types"
)

var (
	ErrAlreadyVoted       = errors.New("vote already exists from voter")
	ErrResolutionNotFound = errors.New("resolution not found")
)

type VoteProcessorStub interface {
	AddVoter(ctx context.Context, identifier []byte, power int64) error
	Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error
	CreateVote(ctx context.Context, body []byte, category string, expiration int64) error
	Expire(ctx context.Context, blockheight int64) error
	GetResolution(ctx context.Context, id types.UUID) (info *ResolutionVoteInfo, err error)
	GetVotesByCategory(ctx context.Context, category string) (votes []*Resolution, err error)
	ProcessConfirmedResolutions(ctx context.Context) error
	RemoveVoter(ctx context.Context, identifier []byte) error
}
