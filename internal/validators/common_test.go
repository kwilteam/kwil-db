package validators

import (
	"bytes"
	"crypto/rand"

	"github.com/kwilteam/kwil-db/core/utils/random"
)

func findValidator(pubkey []byte, vals []*Validator) int {
	for i, v := range vals {
		if bytes.Equal(v.PubKey, pubkey) {
			return i
		}
	}
	return -1
}

var rng = random.New()

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, _ = rand.Read(b)
	return b
}

func newValidator() *Validator {
	return &Validator{
		PubKey: randomBytes(32),
		Power:  rng.Int63n(4) + 1, // in {1,2,3,4}
	}
}
