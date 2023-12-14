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
		argon := crypto.DefaultArgon()
		defer argon.Release()
		stretch := argon.Stretch(password)
		assertions.NotEqual(password, stretch)
		assertions.Len(stretch, crypto.KeySize)
	})
	t.Run("Ensure deterministic", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		password := []byte("password")
		var argon = crypto.DefaultArgon()
		defer argon.Release()
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
		defer argon1.Release()
		var argon2 = crypto.DefaultArgon()
		defer argon2.Release()

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
		defer argon.Release()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.Nil(err, "No error should be returned: %v", err)

		assertions.NotContains(secret.Cipher, k.Buffer[:k.Length])
		assertions.NotContains(secret.Cipher, s.Buffer[:s.Length])
	})
	t.Run("Empty Key", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		k := crypto.RandomData()
		defer k.Release()
		k.Length = 0
		s := crypto.RandomData()
		defer s.Release()
		s.Length = 0
		argon := crypto.DefaultArgon()
		defer argon.Release()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.Nil(err)
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
		defer argon.Release()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.Nil(err)
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
		defer argon.Release()
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
		defer argon.Release()
		secret, err := crypto.Encrypt(k, s, argon)
		defer secret.Release()
		assertions.ErrorIs(err, crypto.ErrNotPrintable)
	})
}

func Test_Decrypt(t *testing.T) {
	encrypt := func(assertions *assert.Assertions) (key, msg crypto.Data, secret crypto.Secret) {
		key = crypto.RandomData()
		msg = crypto.RandomData()
		argon := crypto.DefaultArgon()
		defer argon.Release()
		secret, err := crypto.Encrypt(key, msg, argon)
		assertions.Nil(err, "No error should be returned: %v", err)

		assertions.NotContains(secret.Cipher, key.Buffer[:key.Length])
		assertions.NotContains(secret.Cipher, msg.Buffer[:msg.Length])

		return key, msg, secret
	}
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		key, msg, secret := encrypt(assertions)
		defer key.Release()
		defer msg.Release()
		defer secret.Release()

		result, err := crypto.Decrypt(key, secret)
		assertions.Nil(err)

		assertions.Equal(msg.Bytes(), result.Bytes())

	})
	t.Run("Invalid password", func(t *testing.T) {
		t.Parallel()
		assertions := assert.New(t)

		key, msg, secret := encrypt(assertions)
		key.Release()
		defer msg.Release()
		defer secret.Release()

		result, err := crypto.Decrypt(crypto.RandomData(), secret)
		assertions.NotNil(err)

		assertions.NotEqual(msg.Bytes(), result.Bytes())

	})
}
