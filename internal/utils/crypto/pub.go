package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
)

func NewPrivKey() (key ed25519.PrivateKey, err error) {
	_, key, err = ed25519.GenerateKey(rand.Reader)
	return
}
