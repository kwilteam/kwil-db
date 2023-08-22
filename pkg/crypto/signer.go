package crypto

type Signer interface {
	SignMsg(msg []byte) (*Signature, error)
	PubKey() PublicKey
}

type ComebftSecp256k1Signer struct {
	key *Secp256k1PrivateKey
}

func (c *ComebftSecp256k1Signer) PublicKey() PublicKey {
	return c.key.PubKey()
}

func (c *ComebftSecp256k1Signer) SignMsg(msg []byte) (*Signature, error) {
	hash := Sha256(msg)
	sig, err := c.key.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig[:len(sig)-1],
		Type:      SignatureTypeSecp256k1Cometbft,
	}, nil
}
