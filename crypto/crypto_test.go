package crypto_test

import (
	"crypto/rand"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/stretchr/testify/assert"
)

func TestArgon_Stretch(t *testing.T) {
	t.Run("Ensure diff", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		password := []byte("password")
		stretch := crypto.DefaultArgon().Stretch(password)
		assertions.NotEqual(password, stretch)
		assertions.Len(stretch, crypto.KeySize)
	})
	t.Run("Ensure deterministic", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		password := []byte("password")
		var argon = crypto.DefaultArgon()
		stretch := argon.Stretch(password)
		assertions.NotEqual(password, stretch)
		assertions.Len(stretch, crypto.KeySize)

		assertions.Equal(stretch, argon.Stretch(password))
	})
	t.Run("Ensure uniqueness", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		password := []byte("password")
		var argon1 = crypto.DefaultArgon()
		var argon2 = crypto.DefaultArgon()

		assertions.NotEqual(argon1.Stretch(password), argon2.Stretch(password))
	})
}

func TestEncryption_Release(t *testing.T) {
	t.Run("Ensure safety", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		var secret crypto.Secret
		rand.Read(secret.Cipher[:])
		rand.Read(secret.IV[:])
		secret.KeyArgon = crypto.DefaultArgon()
		secret.ChecksumArgon = crypto.DefaultArgon()
		rand.Read(secret.Checksum[:])

		// Backup
		c := secret.Cipher
		iv := secret.IV
		ka := secret.KeyArgon
		ca := secret.ChecksumArgon
		cs := secret.Checksum

		secret.Release()

		assertions.NotEqual(c, secret.Cipher)
		assertions.NotEqual(iv, secret.IV)
		assertions.NotEqual(ka, secret.KeyArgon)
		assertions.NotEqual(ca, secret.ChecksumArgon)
		assertions.NotEqual(cs, secret.Checksum)
	})
}

func TestData_Release(t *testing.T) {
	t.Run("Ensure safety", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		var d = crypto.RandomData()
		b := d.Buffer
		d.Release()
		assertions.NotEqual(b, d.Buffer)
	})
}

func TestJob_Encrypt(t *testing.T) {
	t.Run("Totally different", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		s := crypto.RandomData()
		defer s.Release()
		argon := crypto.DefaultArgon()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.Nil(err, "No error should be returned: %v", err)

		assertions.NotContains(secret.Cipher, k.Buffer[:k.Length])
		assertions.NotContains(secret.Cipher, s.Buffer[:s.Length])
	})
	t.Run("Longer Secret", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		s := crypto.RandomData()
		defer s.Release()
		s.Length = crypto.ChunkSize
		argon := crypto.DefaultArgon()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.ErrorIs(err, crypto.ErrInvalidSecretSize)
	})
	t.Run("Empty Secret", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		s := crypto.RandomData()
		defer s.Release()
		s.Length = 0
		argon := crypto.DefaultArgon()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.ErrorIs(err, crypto.ErrInvalidSecretSize)
	})
	t.Run("Non printable Key", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		k.Buffer[0] = 0 // null
		s := crypto.RandomData()
		defer s.Release()
		argon := crypto.DefaultArgon()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.ErrorIs(err, crypto.ErrNotPrintable)
	})
	t.Run("Non printable Secret", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		s := crypto.RandomData()
		defer s.Release()
		s.Buffer[0] = 0 // null
		argon := crypto.DefaultArgon()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.ErrorIs(err, crypto.ErrNotPrintable)
	})
}
