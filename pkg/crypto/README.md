This package contains cryptographic functions and types used in Kwil.

Users of this package are expected to use:
- `Signer` to sign message
- `Signature` to verify message

## Signer

A concrete `Signer` knows how to sign the original message according to the signing schema.

Currently, Kwil supports the following signing schemas:
- `ComebftSecp256k1Signer`, cometbft secp256k1 signing schema
- `EthPersonalSecp256k1Signer`, ethereum secp256k1 personal_sign signing schema
- `StdEd25519Signer`, standard ed25519 signing schema

## Signature

A `Signature` is structure that contains the signature and the signing schema type.

Currently, Kwil supports the following signing schemas:
- `secp256k1_ct`, cometbft secp256k1 signing schema
- `secp256k1_ep`, ethereum secp256k1 personal_sign signing schema
- `ed25519`, standard ed25519 signing schema

## Usage

Here we use ed25519 as example:

```go
// sign message
pk, err := Ed25519PrivateKeyFromHex(pvKeyHex)
// err handling ..

edSigner := &StdEd25519Signer{
    key: pk,
}

sig, err := edSigner.SignMsg(msg)
// err handling ..


// verify signature
err = sig.Verify(pk.PubKey(), msg)
// err handling ..
```