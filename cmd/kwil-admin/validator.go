package main

import (
	"context"
	"errors"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
)

// kwil-admin validators list
// kwil-admin validators join-status
// kwil-admin validators join
// kwil-admin validators approve
// kwil-admin validators leave
// kwil-admin validators remove

type ValidatorsCmd struct {
	List       *ValListCmd       `arg:"subcommand:list"`
	JoinStatus *ValJoinStatusCmd `arg:"subcommand:join-status"`
	Join       *ValJoinCmd       `arg:"subcommand:join"`
	Approve    *ValApproveCmd    `arg:"subcommand:approve"`
	Leave      *ValLeaveCmd      `arg:"subcommand:leave"`
	Remove     *ValRemoveCmd     `arg:"subcommand:remove"`

	RPCServer    string `arg:"-s,--rpcserver" default:"127.0.0.1:50051" help:"RPC server address"`
	ChainID      string `arg:"-c,--chain" default:"" help:"The Kwil network's chain ID to use for transactions"`
	OutputFormat string `arg:"-o,--output" default:"text" help:"Output format (text|json)"`
}

// NOTE: each of the validator subcommands satisfies the runner interface to be
// directly callable by main rather than through switches.

type ValListCmd struct{}

var _ runner = (*ValListCmd)(nil)

func (vlc *ValListCmd) run(ctx context.Context, a *args) error {
	var resp respValSets
	err := func() error {
		rpcAddr := a.Vals.RPCServer
		options := []client.Option{client.WithTLSCert("")} // TODO: handle cert
		clt, err := client.Dial(ctx, rpcAddr, options...)
		if err != nil {
			return err
		}

		resp.Data, err = clt.CurrentValidators(ctx)
		return err
	}()

	return display.Print(&resp, err, a.Vals.OutputFormat)
}

type ValJoinStatusCmd struct {
	Joiner HexArg `arg:"positional,required" help:"Public hey (hex) of the candidate validator to check for an active join request."`
}

var _ runner = (*ValJoinStatusCmd)(nil)

func (vjsc *ValJoinStatusCmd) run(ctx context.Context, a *args) error {
	var resp respValJoinStatus
	err := func() error {
		rpcAddr := a.Vals.RPCServer
		options := []client.Option{client.WithTLSCert("")} // TODO: handle cert
		clt, err := client.Dial(ctx, rpcAddr, options...)
		if err != nil {
			return err
		}

		resp.Data, err = clt.ValidatorJoinStatus(ctx, vjsc.Joiner)
		if err != nil {
			if errors.Is(err, client.ErrNotFound) {
				return errors.New("no active join request for that validator")
			}
			return err
		}
		return nil
	}()

	return display.Print(&resp, err, a.Vals.OutputFormat)
}

// edSigningClient makes a client using the provided private key as an ed25519
// Signer.
func edSigningClient(ctx context.Context, rpcAddr, chainID string, privKey []byte) (*client.Client, error) {
	edPrivKey, err := crypto.Ed25519PrivateKeyFromBytes(privKey)
	if err != nil {
		return nil, err
	}

	signer := auth.Ed25519Signer{Ed25519PrivateKey: *edPrivKey}

	options := []client.Option{
		client.WithSigner(&signer, chainID),
		client.WithTLSCert(""),
	}
	return client.Dial(ctx, rpcAddr, options...)
}

// valSignedCmd is meant to be embedded in commands that want a private key in
// either a positional arg or a text file.
type valSignedCmd struct {
	PrivKey HexArg `arg:"positional" help:"(Optional) Private key (hexadecimal) of the validator. Mutually exclusive with --key-file."`

	PrivKeyFile string `arg:"-k,--key-file" help:"File containing the private key of the validator."`
}

func (vsc *valSignedCmd) client(ctx context.Context, rpcAddr, chainID string) (*client.Client, error) {
	privKey, err := keyFromBytesOrFile(vsc.PrivKey, vsc.PrivKeyFile)
	if err != nil {
		return nil, err
	}
	return edSigningClient(ctx, rpcAddr, chainID, privKey)
}

type ValJoinCmd struct {
	valSignedCmd
}

var _ runner = (*ValJoinCmd)(nil)

func (vjc *ValJoinCmd) run(ctx context.Context, a *args) error {
	var txHash []byte
	err := func() error {
		rpcAddr, chainID := a.Vals.RPCServer, a.Vals.ChainID
		clt, err := vjc.client(ctx, rpcAddr, chainID)
		if err != nil {
			return err
		}
		txHash, err = clt.ValidatorJoin(ctx)
		return err
	}()

	return display.Print(display.RespTxHash(txHash), err, a.Vals.OutputFormat)
}

type ValApproveCmd struct {
	Joiner HexArg `arg:"positional,required" help:"Public key of the candidate node with an active join request to approve."`
	valSignedCmd
}

var _ runner = (*ValApproveCmd)(nil)

func (vac *ValApproveCmd) run(ctx context.Context, a *args) error {
	var txHash []byte
	err := func() error {
		rpcAddr, chainID := a.Vals.RPCServer, a.Vals.ChainID
		clt, err := vac.client(ctx, rpcAddr, chainID)
		if err != nil {
			return err
		}
		txHash, err = clt.ApproveValidator(ctx, vac.Joiner)
		return err
	}()

	return display.Print(display.RespTxHash(txHash), err, a.Vals.OutputFormat)
}

type ValRemoveCmd struct {
	Target HexArg `arg:"positional,required" help:"Public key of the validator to propose to remove."`
	valSignedCmd
}

var _ runner = (*ValRemoveCmd)(nil)

func (vac *ValRemoveCmd) run(ctx context.Context, a *args) error {
	var txHash []byte
	err := func() error {
		rpcAddr, chainID := a.Vals.RPCServer, a.Vals.ChainID
		clt, err := vac.client(ctx, rpcAddr, chainID)
		if err != nil {
			return err
		}
		txHash, err = clt.RemoveValidator(ctx, vac.Target)
		return err
	}()

	return display.Print(display.RespTxHash(txHash), err, a.Vals.OutputFormat)
}

type ValLeaveCmd struct {
	valSignedCmd
}

var _ runner = (*ValLeaveCmd)(nil)

func (vjc *ValLeaveCmd) run(ctx context.Context, a *args) error {
	var txHash []byte
	err := func() error {
		rpcAddr, chainID := a.Vals.RPCServer, a.Vals.ChainID
		clt, err := vjc.client(ctx, rpcAddr, chainID)
		if err != nil {
			return err
		}
		txHash, err = clt.ValidatorLeave(ctx)
		return err
	}()

	return display.Print(display.RespTxHash(txHash), err, a.Vals.OutputFormat)
}
