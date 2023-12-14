package crypto_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"log"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
)

func TestArgon_Stretch(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		password := []byte("password")
		argon := crypto.DefaultArgon()
		argon.Salt[0] = 0 // Ensure deterministic results
		defer argon.Release()

		// Verify diff
		stretch := argon.Stretch(password)
		if bytes.Equal(password, stretch) {
			t.Fatalf("Expecting password be completly different from stretched password")
		}

		// Veryfy length
		if len(stretch) != crypto.KeySize {
			t.Fatalf("Expecting length of stretched password be equal to: %d", crypto.KeySize)
		}

		// Verify deterministic
		if !bytes.Equal(stretch, argon.Stretch(password)) {
			t.Fatalf("Expecing a deterministic output")
		}

		// Verify uniqueness
		argon2 := argon
		argon2.Salt[0] = 1
		if bytes.Equal(stretch, argon2.Stretch(password)) {
			t.Fatalf("Expecting stretched passwords be different even by small changes on the salt")
		}

	})
}

func TestEncryption_Release(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		var secret crypto.Secret
		rand.Read(secret.Cipher[:])
		rand.Read(secret.IV[:])
		secret.KeyArgon = crypto.DefaultArgon()
		rand.Read(secret.Checksum[:])
		secret.ChecksumArgon = crypto.DefaultArgon()

		// Backup values
		backup := secret

		// Trigger the release
		secret.Release()

		// Compare
		if backup.IV == secret.IV ||
			backup.Cipher == secret.Cipher ||
			backup.KeyArgon == secret.KeyArgon ||
			backup.Checksum == secret.Checksum ||
			backup.ChecksumArgon == secret.ChecksumArgon {
			t.Fatalf("Expecting values to be different after release: %v == %v", backup, secret)
		}
	})
}

func TestData_Release(t *testing.T) {
	t.Run("Ensure safety", func(t *testing.T) {
		t.Parallel()

		data := crypto.RandomData()
		backup := data
		data.Release()

		if backup == data {
			t.Fatalf("Expecting values to be different after release: %v == %v", backup, data)
		}
	})
}

func TestJob_Encrypt(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name   string
			Key    crypto.Data
			Secret crypto.Data
			Argon  crypto.Argon
		}

		tests := []Test{
			{"Basic", crypto.RandomData(), crypto.RandomData(), crypto.DefaultArgon()},
			{"Empty Key", crypto.Data{}, crypto.RandomData(), crypto.DefaultArgon()},
			{"Empty Secret", crypto.RandomData(), crypto.Data{}, crypto.DefaultArgon()},
		}

		for _, test := range tests {
			test := test
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				secret, err := crypto.Encrypt(test.Key, test.Secret, test.Argon)
				defer secret.Release()
				if err != nil {
					t.Fatalf("No errors expected; received: %v", err)
				}
			})
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name   string
			Key    crypto.Data
			Secret crypto.Data
			Argon  crypto.Argon
			Error  error
		}

		tests := []Test{
			{"Non printable key", crypto.Data{Length: 1}, crypto.RandomData(), crypto.DefaultArgon(), crypto.ErrNotPrintable},
			{"Non printable secret", crypto.RandomData(), crypto.Data{Length: 1}, crypto.DefaultArgon(), crypto.ErrNotPrintable},
		}

		for _, test := range tests {
			test := test
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				secret, err := crypto.Encrypt(test.Key, test.Secret, test.Argon)
				defer secret.Release()
				if !errors.Is(err, test.Error) {
					t.Fatalf("Expecting error to be: %v", err)
				}
			})
		}
	})
}

func Test_Decrypt(t *testing.T) {
	encrypt := func(t *testing.T, key, data crypto.Data) (secret crypto.Secret) {
		argon := crypto.DefaultArgon()
		defer argon.Release()
		secret, err := crypto.Encrypt(key, data, argon)

		if err != nil {
			t.Fatalf("Expecting no errors: %v", err)
		}

		return secret
	}
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name string
			Key  crypto.Data
			Data crypto.Data
		}

		tests := []Test{
			{"Decryption succeed", crypto.RandomData(), crypto.RandomData()},
		}

		for _, test := range tests {
			test := test

			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				secret := encrypt(t, test.Key, test.Data)

				result, err := crypto.Decrypt(test.Key, secret)
				if err != nil {
					t.Fatalf("Expecting no errors: %v", err)
				}

				if !bytes.Equal(test.Data.Bytes(), result.Bytes()) {
					t.Fatalf("Decryption failed decrypted msg doesn't match original msg")
				}
			})
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name          string
			EncryptionKey crypto.Data
			DecryptionKey crypto.Data
			Data          crypto.Data
			Error         error
		}

		tests := []Test{
			{"Invalid Pasword", crypto.RandomData(), crypto.RandomData(), crypto.RandomData(), crypto.ErrInvalidPassword},
		}

		for _, test := range tests {
			test := test

			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				secret := encrypt(t, test.EncryptionKey, test.Data)
				defer secret.Release()

				_, err := crypto.Decrypt(test.DecryptionKey, secret)
				if !errors.Is(err, test.Error) {
					log.Fatalf("Expecting error %v: but received: %v", test.Error, err)
				}
			})
		}
	})
}
