package transactions

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/serialize"

	gethTypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// MsgDescriptionMaxLength is the max length of Description filed in
// TransactionBody and CallMessageBody
const MsgDescriptionMaxLength = 200

const txMsgToSignTmplV0 = `%s

PayloadType: %s
PayloadDigest: %x
Fee: %s
Nonce: %d
Salt: %x

Kwil 🖋
`

// SignedMsgSerializationType is the type of serialization performed on a
// transaction body(in signing and verification)
// The main reason we need this is that this type could also to used as the
// 'version' of the serialization.
// For now, i think it's a bit redundant. To sign a transaction, you need
// three types:
//  1. the type of payload
//  2. the type of serialization
//  3. the type of signature(e.g. signer)
//
// But in the future, take eth signing for example, we might change the
// `signedMsgTmpl` for personal_sign, or `domain` for eip712, this type could
// be used to distinguish the different versions.
//
// NOTE:
// The valid combination of 2.) and 3.) are:
//   - `SignedMsgConcat` + `PersonalSigner/CometBftSigner/NearSigner`, which is
//     the default for the `client` package
//   - `SignedMsgEip712` + `Eip712Signer`
type SignedMsgSerializationType string

func (s SignedMsgSerializationType) String() string {
	return string(s)
}

const (
	// SignedMsgConcat is a human-readable serialization of the transaction body
	// it needs a signer that signs
	SignedMsgConcat SignedMsgSerializationType = "concat"
	// SignedMsgEip712 is specific serialization for EIP712
	SignedMsgEip712 SignedMsgSerializationType = "eip712"

	// DefaultSignedMsgSerType is the default serialization type
	// It's `concat` for now, since it's the only one known works for every signer
	DefaultSignedMsgSerType = SignedMsgConcat
)

const (
	// EIP712DomainTypeName TODO: to test if we can change this
	EIP712DomainTypeName = "EIP712Domain"
)

// EIP712TypedDomain represents the domain separator for EIP712
var EIP712TypedDomain = []gethTypes.Type{
	{Name: "name", Type: "string"},
	{Name: "version", Type: "string"},
	//{Name: "chainId", Type: "uint256"},
	//{Name: "verifyingContract", Type: "address"},
	// NOTE: Domain.Salt is different from TransactionBody.Salt,
	// Domain.Salt is last resort to distinguish different Dapp
	{Name: "salt", Type: "string"},
}

// EIP712TypedDataMessage represents the primary type for EIP712 data
var EIP712TypedDataMessage = []gethTypes.Type{
	// type is the type of the payload
	{Name: "payload_type", Type: "string"},
	// data is the payload, which is a JSON string
	{Name: "payload_digest", Type: "string"},
	// fee is the fee for the transaction
	{Name: "fee", Type: "string"},
	// nonce is the nonce for the transaction
	{Name: "nonce", Type: "string"},
	// salt is the salt for the transaction
	{Name: "salt", Type: "string"},
}

// CreateTransaction creates a new unsigned transaction.
func CreateTransaction(contents Payload, nonce uint64) (*Transaction, error) {
	data, err := contents.MarshalBinary()
	if err != nil {
		return nil, err
	}

	salt, err := generateRandomSalt()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Body: &TransactionBody{
			Payload:     data,
			PayloadType: contents.Type(),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
			Salt:        salt[:],
		},
		Serialization: DefaultSignedMsgSerType,
	}, nil
}

type Transaction struct {
	// Signature is the signature of the transaction.
	Signature *auth.Signature

	// Body is the body of the transaction. It gets serialized and signed.
	Body *TransactionBody

	// Serialization is the serialization performed on `Body`
	// in order to generate the message that being signed.
	Serialization SignedMsgSerializationType

	// Sender is the public key of the sender
	// NOTE: we could not use crypto.PublicKey, since it's not RLP-serializable
	Sender []byte
}

// SerializeMsg produces the serialization of the transaction that is to be used
// in both signing and verification of transaction.
func (t *Transaction) SerializeMsg() ([]byte, error) {
	return t.Body.SerializeMsg(t.Serialization)
}

// Verify verifies the authenticity of a signed transaction. It will serialize
// the transaction body to get message-to-be-signed, then verify it.
//
// NOTE: This is not the recommended way to verify a transaction in kwild since
// it should use its internal VerifyTransaction function that uses its
// Authenticator registry. This method is provided so SDK users may verify
// transactions directly with an Verifier implementation, such as those provided
// in core/crypto/auth.
func (t *Transaction) Verify(verifier auth.Verifier) error {
	msg, err := t.Body.SerializeMsg(t.Serialization)
	if err != nil {
		return err
	}

	return verifier.Verify(t.Sender, msg, t.Signature.Signature)
}

// Sign signs transaction body with given signer.
// It will serialize the transaction body first and sign it.
func (t *Transaction) Sign(signer auth.Signer) error {
	msg, err := t.Body.SerializeMsg(t.Serialization)
	if err != nil {
		return err
	}

	signature, err := signer.Sign(msg)
	if err != nil {
		return err
	}

	t.Signature = signature
	t.Sender = signer.PublicKey()

	return nil
}

func (t *Transaction) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(t)
}

func (t *Transaction) UnmarshalBinary(data serialize.SerializedData) error {
	return serialize.DecodeInto(data, t)
}

// TransactionBody is the body of a transaction that gets included in the signature
// NOTE: rlp encoding will preserve the order of the fields
type TransactionBody struct {
	// Description is a human-readable description of the transaction
	Description string

	// Payload is the raw bytes of the payload data, it is RLP encoded
	Payload serialize.SerializedData

	// PayloadType is the type of the payload
	// This can be used to determine how to decode the payload
	PayloadType PayloadType

	// Fee is the fee the sender is willing to pay for the transaction
	Fee *big.Int

	// Nonce is the next nonce of the sender
	Nonce uint64

	// Salt is a random value that is used to prevent replay attacks and hash collisions
	Salt []byte
}

func (t *TransactionBody) MarshalBinary() ([]byte, error) {
	return serialize.Encode(t)
}

// SerializeMsg prepares a message for signing or verification using a certain
// message construction format. This is done since a Kwil transaction is foreign
// to wallets, and it is signed as a message, not a transaction that is native
// to the wallet. As such we define conventions for constructing user-friendly
// messages. The Kwil frontend SDKs much implement these serialization schemes.
func (t *TransactionBody) SerializeMsg(mst SignedMsgSerializationType) ([]byte, error) {
	if len(t.Description) > MsgDescriptionMaxLength {
		return nil, errors.New("description is too long")
	}

	switch mst {
	case SignedMsgConcat:
		// Make a human-readable message using a template(txMsgToSignTmplV0).
		// In this message scheme, the displayed "token" is a hash of the
		// payload.
		// NOTE: 'payload` is still in binary form(RLP encoded),
		// we present its hash in the result message.
		payloadDigest := crypto.Sha256(t.Payload)[:20]
		msgStr := fmt.Sprintf(txMsgToSignTmplV0,
			t.Description,
			t.PayloadType.String(),
			payloadDigest,
			t.Fee.String(),
			t.Nonce,
			t.Salt)
		return []byte(msgStr), nil
		//case SignedMsgEip712:
		//	signerData := gethTypes.TypedData{
		//		Types: gethTypes.Types{
		//			EIP712DomainTypeName: EIP712TypedDomain,
		//			"Message":            EIP712TypedDataMessage,
		//		},
		//		PrimaryType: "Message",
		//		Domain: gethTypes.TypedDataDomain{
		//			Name:    "Kwil", // TODO: should this be the name of the Dapp?
		//			Version: "1",
		//			// NOTE: not sure what should be treated as the Dapp on kwil
		//			// either kwil itself or the Kuneiform
		//			// if Kuneiform, DB_ID could be the salt?
		//			Salt: hex.EncodeToString(t.Salt),
		//		},
		//		Message: gethTypes.TypedDataMessage{
		//			"payload_type":   t.PayloadType.String(),
		//			"payload_digest": string(t.Payload),
		//			"fee":            t.Fee,
		//			"nonce":          t.Nonce,
		//			"salt":           hex.EncodeToString(t.Salt),
		//		},
		//	}
		//
		//	return eip712TypedDataAndHash(signerData)

	}
	return nil, errors.New("invalid serialization type")
}

// generateRandomSalt generates a new random salt
// this salt is not used for any sort of security purpose;
// rather, it is just to prevent hash collisions
// therefore, we only need a small amount of entropy
func generateRandomSalt() ([8]byte, error) {
	var s [8]byte

	_, err := rand.Read(s[:])
	if err != nil {
		return s, err
	}
	return s, nil
}

// TxHash is the hash of a transaction that could be used to query the transaction
type TxHash []byte

func (h TxHash) Hex() string {
	return strings.ToUpper(fmt.Sprintf("%x", h))
}

//// eip712TypedDataAndHash returns the `hashStruct(eip712Domain) || hashStruct(message)`
//// of the given TypedData.
//// It's different from the function `TypedDataAndHash` in go-ethereum signer
//// package in that it only return `<version specific data> <data to sign>`.
//func eip712TypedDataAndHash(typedData gethTypes.TypedData) ([]byte, error) {
//	domainSeparator, err := typedData.HashStruct(EIP712DomainTypeName, typedData.Domain.Map())
//	if err != nil {
//		return nil, err
//	}
//
//	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
//	if err != nil {
//		return nil, err
//	}
//
//	rawData := fmt.Sprintf("%s%s", string(domainSeparator), string(typedDataHash))
//	return []byte(rawData), nil
//}
