package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

// kwil-admin validators list
// kwil-admin validators join-status
// kwil-admin validators join
// kwil-admin validators approve
// kwil-admin validators leave

type ValidatorsCmd struct {
	List       *ValListCmd       `arg:"subcommand:list"`
	JoinStatus *ValJoinStatusCmd `arg:"subcommand:join-status"`
	Join       *ValJoinCmd       `arg:"subcommand:join"`
	Approve    *ValApproveCmd    `arg:"subcommand:approve"`
	Leave      *ValLeaveCmd      `arg:"subcommand:leave"`

	RPCServer string `arg:"-s,--rpcserver" default:"127.0.0.1:50051" help:"RPC server address"`
}

// NOTE: each of the validator subcommands satisfies the runner interface to be
// directly callable by main rather than through switches.

type ValListCmd struct{}

var _ runner = (*ValListCmd)(nil)

func (vlc *ValListCmd) run(ctx context.Context, a *args) error {
	rpcAddr := a.Vals.RPCServer
	options := []client.ClientOpt{}
	clt, err := client.New(rpcAddr, options...)
	if err != nil {
		return err
	}

	vals, err := clt.CurrentValidators(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Current validator set:")
	for i, v := range vals {
		fmt.Printf("% 3d. %v\n", i, v)
	}
	return nil
}

type ValJoinStatusCmd struct {
	Joiner HexArg `arg:"positional,required" help:"Public hey (hex) of the candidate valiator to check for an active join request."`
}

var _ runner = (*ValJoinStatusCmd)(nil)

func (vjsc *ValJoinStatusCmd) run(ctx context.Context, a *args) error {
	rpcAddr := a.Vals.RPCServer
	options := []client.ClientOpt{}
	clt, err := client.New(rpcAddr, options...)
	if err != nil {
		return err
	}

	status, err := clt.ValidatorJoinStatus(ctx, vjsc.Joiner)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			fmt.Println("No active join request for that validator.")
			return nil
		}
		return err
	}
	fmt.Printf("Candidate: %v (want power %d)\n", base64.StdEncoding.EncodeToString(status.Candidate), status.Power)
	for i := range status.Board {
		fmt.Printf(" Validator %x, approved = %v\n", status.Board[i], status.Approved[i])
	}
	return nil
}

// edSigningClient makes a client using the provided private key as an ed25519
// Signer.
func edSigningClient(rpcAddr string, privKey []byte) (*client.Client, error) {
	edPrivKey, err := crypto.Ed25519PrivateKeyFromBytes(privKey)
	if err != nil {
		return nil, err
	}

	signer := crypto.NewStdEd25519Signer(edPrivKey)
	options := []client.ClientOpt{client.WithSigner(signer)}
	return client.New(rpcAddr, options...)
}

// valSignedCmd is meant to be embedded in commands that want a private key in
// either a positional arg or a text file.
type valSignedCmd struct {
	PrivKey HexArg `arg:"positional" help:"Private key (hexadecimal string) of the validator"`

	PrivKeyFile string `arg:"-k,--key-file" help:"file containing the private key of the validator"`
}

func (vsc *valSignedCmd) client(rpcAddr string) (*client.Client, error) {
	privKey, err := keyFromBytesOrFile(vsc.PrivKey, vsc.PrivKeyFile)
	if err != nil {
		return nil, err
	}
	return edSigningClient(rpcAddr, privKey)
}

type ValJoinCmd struct {
	valSignedCmd
}

var _ runner = (*ValJoinCmd)(nil)

func (vjc *ValJoinCmd) run(ctx context.Context, a *args) error {
	rpcAddr := a.Vals.RPCServer
	clt, err := vjc.client(rpcAddr)
	if err != nil {
		return err
	}
	hash, err := clt.ValidatorJoin(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Join transaction hash: %x\n", hash)
	return nil
}

type ValApproveCmd struct {
	Joiner HexArg `arg:"positional,required"`
	valSignedCmd
}

var _ runner = (*ValApproveCmd)(nil)

func (vac *ValApproveCmd) run(ctx context.Context, a *args) error {
	rpcAddr := a.Vals.RPCServer
	clt, err := vac.client(rpcAddr)
	if err != nil {
		return err
	}
	hash, err := clt.ApproveValidator(ctx, vac.Joiner)
	if err != nil {
		return err
	}
	fmt.Printf("Approval transaction hash: %x\n", hash)
	return nil
}

type ValLeaveCmd struct {
	valSignedCmd
}

var _ runner = (*ValLeaveCmd)(nil)

func (vjc *ValLeaveCmd) run(ctx context.Context, a *args) error {
	rpcAddr := a.Vals.RPCServer
	clt, err := vjc.client(rpcAddr)
	if err != nil {
		return err
	}
	hash, err := clt.ValidatorLeave(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Leave transaction hash: %x\n", hash)
	return nil
}
