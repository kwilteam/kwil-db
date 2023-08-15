package crypto

type Signer interface {
	SignMsg(msg []byte) (*Signature, error)
}

type Eip712Signer struct {
	key PrivateKey
}

func (e *Eip712Signer) SignMsg(msg []byte) (*Signature, error) {
	panic("not implemented")
}

type ComebftSecp256k1Signer struct {
	key PrivateKey
}

func (c *ComebftSecp256k1Signer) SignMsg(msg []byte) (*Signature, error) {
	hash := Sha256(msg)
	sig, err := c.key.Sign(hash)
	if err != nil {
		return nil, err
	}
	return &Signature{
		Signature: sig[:len(sig)-1],
		Type:      SIGNATURE_TYPE_SECP256K1_COMETBFT,
	}, nil
}
