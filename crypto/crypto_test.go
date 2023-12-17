package crypto_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/internal/testsuite"
)

func TestJob_Encrypt_Encrypt(t *testing.T) {
	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name string
			crypto.Job
		}

		tests := []Test{
			{"Basic", crypto.Job{testsuite.Random(16), testsuite.Random(16), crypto.DefaultArgon(), 16}},
			{"Empty Key", crypto.Job{make([]byte, 16), testsuite.Random(16), crypto.DefaultArgon(), 16}},
			{"Empty Data", crypto.Job{testsuite.Random(16), make([]byte, 16), crypto.DefaultArgon(), 16}},
		}

		for _, test := range tests {
			test := test
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				// Encrypt
				encryption := crypto.Job{
					Key:      test.Key,
					Data:     test.Data,
					Argon:    test.Argon,
					SaltSize: test.SaltSize,
				}
				defer encryption.Release()
				secret := encryption.Encrypt()
				defer secret.Release()

				// Decrypt
				// Verify decrypted data matches
				decryption := crypto.Job{
					Key:      test.Key,
					Argon:    test.Argon,
					SaltSize: test.SaltSize,
				}
				defer decryption.Release()

				err := decryption.Decrypt(secret)
				if err != nil {
					t.Fatalf("expecting no error but obtained: %v", err)
				}

				// Compare results
				if !bytes.Equal(encryption.Data, decryption.Data) {
					t.Logf("%s != %s", hex.EncodeToString(encryption.Data), hex.EncodeToString(decryption.Data))
					t.Fatal("expecting decryption result be equal to encryption result")
				}
			})
		}
	})
}
