package types

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/big"

	"kwil/crypto"
	"kwil/crypto/auth"
)

// MsgDescriptionMaxLength is the max length of Description filed in
// TransactionBody and CallMessageBody
const MsgDescriptionMaxLength = 200

type Transaction struct {
	// Signature is the signature of the transaction.
	Signature *auth.Signature `json:"signature,omitempty"`

	// Body is the body of the transaction. It gets serialized and signed.
	Body *TransactionBody `json:"body,omitempty"`

	// Serialization is the serialization performed on `Body`
	// in order to generate the message that being signed.
	Serialization SignedMsgSerializationType `json:"serialization"`

	// Sender is the user identifier, which is generally an address but may be
	// a public key of the sender.
	Sender HexBytes `json:"sender"`

	strictUnmarshal bool
}

func (t *Transaction) StrictUnmarshal() {
	t.strictUnmarshal = true
}

// TransactionBody is the body of a transaction that gets included in the
// signature. This type implements json.Marshaler and json.Unmarshaler to ensure
// that the Fee field is represented as a string in JSON rather than a number.
type TransactionBody struct {
	// Description is a human-readable description of the transaction.
	Description string `json:"desc"`

	// Payload is the raw bytes of the payload data.
	Payload []byte `json:"payload"`

	// PayloadType is the type of the payload, which may be used to determine
	// how to decode the payload.
	PayloadType PayloadType `json:"type"`

	// Fee is the fee the sender is willing to pay for the transaction.
	Fee *big.Int `json:"fee"` // MarshalJSON and UnmarshalJSON handle this field, but still tagged for reflection

	// Nonce should be the next nonce of the sender..
	Nonce uint64 `json:"nonce"`

	// ChainID identifies the Kwil chain for which the transaction is intended.
	// Alternatively, this could be withheld from the TransactionBody and passed
	// as an argument to SerializeMsg, as is seen in ethereum signers. However,
	// the full transaction serialization must include it anyway since it passes
	// through the consensus engine and p2p systems as an opaque blob that must
	// be unmarshaled with the chain ID in Kwil blockchain application.
	ChainID string `json:"chain_id"`

	strictUnmarshal bool
}

func (tb *TransactionBody) StrictUnmarshal() {
	tb.strictUnmarshal = true
}

const txMsgToSignTmplV0 = `%s

PayloadType: %s
PayloadDigest: %x
Fee: %s
Nonce: %d

Kwil Chain ID: %s
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

	// DefaultSignedMsgSerType is the default serialization type
	// It's `concat` for now, since it's the only one known works for every signer
	DefaultSignedMsgSerType = SignedMsgConcat
)

// CreateTransaction creates a new unsigned transaction.
func CreateTransaction(contents Payload, chainID string, nonce uint64) (*Transaction, error) {
	data, err := contents.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Body: &TransactionBody{
			Payload:     data,
			PayloadType: contents.Type(),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
			ChainID:     chainID,
		},
		Serialization: DefaultSignedMsgSerType,
	}, nil
}

// SerializeMsg produces the serialization of the transaction that is to be used
// in both signing and verification of transaction.
func (t *Transaction) SerializeMsg() ([]byte, error) {
	return t.Body.SerializeMsg(t.Serialization)
}

// Sign signs transaction body with given signer.
// It will serialize the transaction body first and sign it.
func (t *Transaction) Sign(signer auth.Signer) error {
	msg, err := t.SerializeMsg()
	if err != nil {
		return err
	}
	// The above serialized msg has to include the chainID rather than passing
	// it to the signer because it needs to be displayed in the friendly message
	// that the user signs.

	signature, err := signer.Sign(msg)
	if err != nil {
		return err
	}

	t.Signature = signature
	t.Sender = signer.Identity()

	return nil
}

// SerializeMsg prepares a message for signing or verification using a certain
// message construction format. This is done since a Kwil transaction is foreign
// to wallets, and it is signed as a message, not a transaction that is native
// to the wallet. As such we define conventions for constructing user-friendly
// messages. The Kwil frontend SDKs must implement these serialization schemes.
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
			t.ChainID)
		return []byte(msgStr), nil
	}
	return nil, errors.New("invalid serialization type")
}

// MarshalBinary produces the full binary serialization of the transaction,
// which is the form used in p2p messaging and blockchain storage.
func (t *Transaction) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := t.serialize(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ io.ReaderFrom = (*Transaction)(nil)

func (t *Transaction) ReadFrom(r io.Reader) (int64, error) {
	n, err := t.deserialize(r)
	if err != nil {
		return int64(n), err
	}
	return int64(n), nil
}

func (t *Transaction) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	n, err := t.deserialize(r)
	if err != nil {
		return err
	}
	if !t.strictUnmarshal {
		return nil
	}
	if n != len(data) {
		return errors.New("failed to read all")
	}
	if r.Len() != 0 {
		return errors.New("extra transaction data")
	}
	return nil
}

func (tb *TransactionBody) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Description Length + Description
	if err := writeString(buf, tb.Description); err != nil {
		return nil, fmt.Errorf("failed to write transaction body description: %w", err)
	}

	// serialized Payload
	if err := writeBytes(buf, tb.Payload); err != nil {
		return nil, fmt.Errorf("failed to write transaction body payload: %w", err)
	}

	// PayloadType
	payloadType := tb.PayloadType.String()
	if err := writeString(buf, payloadType); err != nil {
		return nil, fmt.Errorf("failed to write transaction body payload type: %w", err)
	}

	// Fee (big.Int)
	if err := writeBigInt(buf, tb.Fee); err != nil {
		return nil, fmt.Errorf("failed to write transaction fee: %w", err)
	}

	// Nonce
	if err := binary.Write(buf, binary.LittleEndian, tb.Nonce); err != nil {
		return nil, fmt.Errorf("failed to write transaction body nonce: %w", err)
	}

	// ChainID
	if err := writeString(buf, tb.ChainID); err != nil {
		return nil, fmt.Errorf("failed to write transaction body chain ID: %w", err)
	}

	return buf.Bytes(), nil
}

var _ io.ReaderFrom = (*TransactionBody)(nil)

func (tb *TransactionBody) ReadFrom(r io.Reader) (int64, error) {
	var n int
	// Description Length + Description
	desc, err := readString(r)
	if err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body description: %w", err)
	}
	tb.Description = desc
	n += 4 + len(desc)

	// serialized Payload
	payload, err := readBytes(r)
	if err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body payload: %w", err)
	}
	tb.Payload = payload
	n += 4 + len(payload)

	// PayloadType
	payloadType, err := readString(r)
	if err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body payload type: %w", err)
	}
	tb.PayloadType = PayloadType(payloadType)
	n += 4 + len(payloadType)

	// Fee (big.Int)
	b, ni, err := readBigInt(r)
	if err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body fee: %w", err)
	}
	tb.Fee = b // may be nil
	n += ni

	// Nonce
	if err := binary.Read(r, binary.LittleEndian, &tb.Nonce); err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body nonce: %w", err)
	}
	n += 8

	// ChainID
	chainID, err := readString(r)
	if err != nil {
		return int64(n), fmt.Errorf("failed to read transaction body chain ID: %w", err)
	}
	tb.ChainID = chainID
	n += 4 + len(chainID)

	return int64(n), nil
}

func (tb *TransactionBody) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	n, err := tb.ReadFrom(buf)
	if err != nil {
		return err
	}

	if !tb.strictUnmarshal {
		return nil
	}

	if int(n) != len(data) {
		return errors.New("extra tx body data")
	}
	if buf.Len() != 0 {
		return errors.New("extra tx body data (buf)")
	}

	return nil
}

func (t *Transaction) serialize(w io.Writer) (err error) {
	// Tx Signature
	var sigBytes []byte
	if t.Signature != nil {
		if sigBytes, err = t.Signature.MarshalBinary(); err != nil {
			return fmt.Errorf("failed to marshal transaction signature: %w", err)
		}
	}
	if err := writeBytes(w, sigBytes); err != nil {
		return fmt.Errorf("failed to write transaction signature: %w", err)
	}

	// Tx Body
	if t.Body == nil {
		return errors.New("missing transaction body")
	}
	txBodyBytes, err := t.Body.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal transaction body: %w", err)
	}
	if _, err := w.Write(txBodyBytes); err != nil {
		return fmt.Errorf("failed to write transaction body: %w", err)
	}
	/*var txBodyBytes []byte
	if t.Body != nil { // why support this?
		txBodyBytes, err = t.Body.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal transaction body: %w", err)
		}
	}
	if err := writeBytes(w, txBodyBytes); err != nil {
		return fmt.Errorf("failed to write transaction body: %w", err)
	}*/

	// SerializationType
	if err := writeString(w, string(t.Serialization)); err != nil {
		return fmt.Errorf("failed to write transaction serialization type: %w", err)
	}

	// Sender
	if err := writeBytes(w, t.Sender); err != nil {
		return fmt.Errorf("failed to write transaction sender: %w", err)
	}

	return nil
}

func (t *Transaction) deserialize(r io.Reader) (int, error) {
	var n int

	// Signature
	sigBytes, err := readBytes(r)
	if err != nil {
		return n, fmt.Errorf("failed to read transaction signature: %w", err)
	}
	n += 4 + len(sigBytes)

	if len(sigBytes) != 0 {
		var signature auth.Signature
		if err = signature.UnmarshalBinary(sigBytes); err != nil {
			return 0, fmt.Errorf("failed to unmarshal transaction signature: %w", err)
		}
		t.Signature = &signature
	}

	// TxBody
	var body TransactionBody
	bodyLen, err := body.ReadFrom(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read transaction body: %w", err)
	}
	t.Body = &body
	n += int(bodyLen)
	/* if we need to support nil body...
	bodyBytes, err := readBytes(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read transaction body: %w", err)
	}
	n += 4 + len(bodyBytes)
	if len(bodyBytes) != 0 {
		var body TransactionBody
		body.StrictUnmarshal()
		if err := body.UnmarshalBinary(bodyBytes); err != nil {
			return 0, fmt.Errorf("failed to unmarshal transaction body: %w", err)
		}
		t.Body = &body
	}*/

	// SerializationType
	serType, err := readString(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read transaction serialization type: %w", err)
	}
	n += 4 + len(serType)
	t.Serialization = SignedMsgSerializationType(serType)

	// Sender
	senderBytes, err := readBytes(r)
	if err != nil {
		return 0, fmt.Errorf("failed to read transaction sender: %w", err)
	}
	n += 4 + len(senderBytes)
	t.Sender = senderBytes

	return n, nil
}

func writeBytes(w io.Writer, data []byte) error {
	if err := binary.Write(w, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err := w.Write(data)
	return err
}

func writeString(w io.Writer, s string) error {
	return writeBytes(w, []byte(s))
}

func readBytes(r io.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	if length == 0 {
		return nil, nil
	}

	var data []byte
	if rl, ok := r.(interface{ Len() int }); ok {
		if int(length) > rl.Len() {
			return nil, fmt.Errorf("encoded length %d is longer than data length %d", length, rl.Len())
		}
		data = make([]byte, length)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
	} else {
		buf := &bytes.Buffer{}
		_, err := io.CopyN(buf, r, int64(length))
		if err != nil {
			return nil, fmt.Errorf("failed to read signature data: %w", err)
		}
		data = buf.Bytes()
	}

	return data, nil
}

func readString(r io.Reader) (string, error) {
	bts, err := readBytes(r)
	return string(bts), err
}

func writeBigInt(w io.Writer, b *big.Int) error {
	if b == nil {
		_, err := w.Write([]byte{0})
		return err
	}

	_, err := w.Write([]byte{1})
	if err != nil {
		return err
	}

	// This is ridiculous, maybe we should just use String() and SetString()
	// var negByte byte
	// if b.Sign() < 0 {
	// 	negByte = 1
	// }
	// _, err = w.Write([]byte{negByte})
	// if err != nil {
	// 	return err
	// }
	// return writeBytes(w, b.Bytes())

	return writeString(w, b.String())
}

func readBigInt(r io.Reader) (*big.Int, int, error) {
	nilByte := []byte{0}
	n, err := io.ReadFull(r, nilByte)
	if err != nil {
		return nil, n, err
	}

	switch nilByte[0] {
	case 0:
		return nil, n, nil
	case 1:
	default:
		return nil, n, errors.New("invalid nil int byte")
	}

	intStr, err := readString(r)
	if err != nil {
		return nil, n, err
	}
	n += 4 + len(intStr)

	b := new(big.Int)
	b, ok := b.SetString(intStr, 10)
	if !ok {
		return nil, n, errors.New("bad big int string")
	}
	if b.String() != intStr {
		return nil, n, errors.New("non-canonical big int encoding")
	}

	// negByte := []byte{0}
	// _, err = io.ReadFull(r, negByte)
	// if err != nil {
	// 	return nil, err
	// }
	// b := new(big.Int).SetBytes(intBts)
	// if negByte[0] == 1 {
	// 	b.Neg(b)
	// }
	// if !bytes.Equal(b.Bytes(), intBts) {
	// 	return nil, errors.New("non-canonical big int encoding")
	// }
	return b, n, nil
}
