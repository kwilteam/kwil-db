package addresses

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/kwilteam/go-sqlite"
	"github.com/kwilteam/kwil-db/internal/engine/types" // this is the one binding to engine in sql/sqlite
	"github.com/kwilteam/kwil-db/internal/ident"
)

func Register(c *sqlite.Conn) error {
	err := c.CreateFunction("address", AddressImpl)
	if err != nil {
		return fmt.Errorf(`failed to register "address" function: %w`, err)
	}

	err = c.CreateFunction("public_key", publicKeyImpl)
	if err != nil {
		return fmt.Errorf(`failed to register "public_key" function: %w`, err)
	}

	return nil
}

const (
	// addressFuncName is the name of the ADDRESS function.
	addressFuncName = "ADDRESS"
	// publicKeyFuncName is the name of the PUBLIC_KEY function.
	publicKeyFuncName = "PUBLIC_KEY"
)

// addressImpl is the implementation of the ADDRESS function.
var AddressImpl = &sqlite.FunctionImpl{
	NArgs:         1,
	Deterministic: true,
	AllowIndirect: true,
	Scalar:        addressFunc,
}

// addressFunc is the function that is called whenever the ADDRESS
// function is invoked.
func addressFunc(ctx sqlite.Context, args []sqlite.Value) (sqlite.Value, error) {
	if len(args) != 1 {
		return raiseErr(addressFuncName, fmt.Errorf("expected 1 argument, got %d", len(args)))
	}

	identifier, err := getIdentifier(args[0])
	if err != nil {
		return raiseErr(addressFuncName, fmt.Errorf("failed to read public key identifier: %w", err))
	}

	address, err := ident.Address(identifier.AuthType, identifier.PublicKey)
	if err != nil {
		return raiseErr(addressFuncName, fmt.Errorf("failed to get address: %w", err))
	}

	return sqlite.TextValue(address), nil
}

// publicKeyImpl is the implementation of the PUBLIC_KEY function.
var publicKeyImpl = &sqlite.FunctionImpl{
	NArgs:         -1, // variadic
	Deterministic: true,
	AllowIndirect: true,
	Scalar:        publicKeyFunc,
}

// publicKeyFunc is the function that is called whenever the PUBLIC_KEY
// function is invoked.
func publicKeyFunc(ctx sqlite.Context, args []sqlite.Value) (sqlite.Value, error) {
	if len(args) > 2 || len(args) < 1 {
		return raiseErr(publicKeyFuncName, fmt.Errorf("expected 1 or 2 arguments, got %d", len(args)))
	}

	ident, err := getIdentifier(args[0])
	if err != nil {
		return raiseErr(publicKeyFuncName, fmt.Errorf("failed to read public key identifier: %w", err))
	}

	// the encodingType to return as
	encodingType := "blob"
	if len(args) == 2 {
		encodingType = args[1].Text()
	}

	switch encodingType {
	default:
		return raiseErr(publicKeyFuncName, fmt.Errorf("invalid encoding type: %s", encodingType))
	case "hex":
		return sqlite.TextValue(hex.EncodeToString(ident.PublicKey)), nil
	case "base64":
		return sqlite.TextValue(base64.StdEncoding.EncodeToString(ident.PublicKey)), nil
	case "base64url":
		return sqlite.TextValue(base64.URLEncoding.EncodeToString(ident.PublicKey)), nil
	case "blob":
		return sqlite.BlobValue(ident.PublicKey), nil
	}
}

// getIdentifier takes in a sqlite input and that it is a blob.
// It then unmarshals the blob into a KeyIdentifier.
// It will check if it is a valid key identifier.
// An example of an invalid key identifier is a key identifier
// that has uses Ethereum addresses and an ed25519 public key.
func getIdentifier(arg sqlite.Value) (*types.User, error) {
	// matches the crypto/addresses/identifiers.go:KeyIdentifier
	userIdentifierBlob := arg.Blob()
	ident := &types.User{}
	err := ident.UnmarshalBinary(userIdentifierBlob)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key identifier: %w", err)
	}

	return ident, nil
}

// raiseErr is a helper function that returns an error with the given
// function name.
func raiseErr(functionName string, err error) (sqlite.Value, error) {
	return sqlite.Value{}, fmt.Errorf(`failed to execute function "%s": %w`, functionName, err)
}
