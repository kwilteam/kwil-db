This package contains cryptographic functions and types used in Kwil.

This package contains two pairs of `sign` and `verify`:
- `PrivateKey.Sign` and `PublicKey.Verify`, this are low-level methods for different wallet, user won't touch it
- `Signer.Sign` and `Signature.Verify`, this are util functions for different standard SigningSchema(for example,
    eip-191), which have different message process or different hashing algorithm etc

There is another one worth mentioning in `pkg/transactions`, although these are not crypto primitives:
- `Transaction.Sign` and `Transaction.Verify`, they share same Sign/Veify interface.
    Those methods will choose different SigningSchema to sign/verify transaction, based on Kwil/dApp's requirement.
    Typically, these are the methods that Kwil/dApp use directly.

## Signer

A concrete `Signer` knows how to sign the original message according to the corresponding signing schema.

Currently, Kwil supports the following signing schemas:
- `ComebftSecp256k1Signer`, cometbft secp256k1 signing schema
- `EthPersonalSecp256k1Signer`, ethereum secp256k1 personal_sign signing schema
- `StdEd25519Signer`, standard ed25519 signing schema
- `NearEd25519Signer`, near ed25519 signing schema

## Signature

A `Signature` is structure that contains the signature and the signing schema type.

Currently, Kwil supports the following signing schemas:
- `secp256k1_ct`, cometbft secp256k1 signing schema
- `secp256k1_ep`, ethereum secp256k1 personal_sign signing schema
- `ed25519`, standard ed25519 signing schema
- `ed25519-nr`, near ed25519 signing schema

## Usage

Here we use ed25519 as example:

```go
// sign message
pk, err := Ed25519PrivateKeyFromHex(pvKeyHex)
// err handling ..

edSigner := &StdEd25519Signer{
    key: pk,
}

sig, err := edSigner.Sign(msg)
// err handling ..


// verify signature
err = sig.Verify(pk.PubKey(), msg)
// err handling ..
```