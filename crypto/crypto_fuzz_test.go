package crypto_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
)

const (
	MaxKeyLength  = 12
	MaxDataLength = 12
)

func Fuzz_Encrypt(f *testing.F) {
	for keyLength := uint8(0); keyLength <= MaxKeyLength; keyLength++ {
		for dataLength := uint8(0); dataLength <= MaxDataLength; dataLength++ {
			f.Add(keyLength, dataLength)
		}
	}
	f.Fuzz(func(t *testing.T, keyLength, dataLength uint8) {
		key := crypto.RandomData()
		defer key.Release()
		key.Length = keyLength

		data := crypto.RandomData()
		defer data.Release()
		data.Length = dataLength

		secret, err := crypto.Encrypt(key, data, crypto.DefaultArgon())
		if err != nil {
			t.Fatalf("Expecting no errors: received: %v", err)
		}
		defer secret.Release()
	})
}

func Fuzz_Decrypt(f *testing.F) {
	for keyLength := uint8(0); keyLength <= MaxKeyLength; keyLength++ {
		for dataLength := uint8(0); dataLength <= MaxDataLength; dataLength++ {
			f.Add(keyLength, dataLength)
		}
	}
	f.Fuzz(func(t *testing.T, keyLength, dataLength uint8) {
		key := crypto.RandomData()
		defer key.Release()
		key.Length = keyLength

		data := crypto.RandomData()
		defer data.Release()
		data.Length = dataLength

		secret, err := crypto.Encrypt(key, data, crypto.DefaultArgon())
		if err != nil {
			t.Fatalf("Expecting no errors on encryption: received: %v", err)
		}
		defer secret.Release()

		decrypted, err := crypto.Decrypt(key, secret)
		if err != nil {
			t.Fatalf("Expecting no errors on decryption: received: %v", err)
		}

		if !bytes.Equal(data.Bytes(), decrypted.Bytes()) {
			t.Fatalf("Decryption failed, inconsistent data: %v != %v", hex.EncodeToString(data.Bytes()), hex.EncodeToString(decrypted.Bytes()))
		}
	})
}
