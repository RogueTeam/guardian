package database_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
)

func TestJson(t *testing.T) {
	t.Parallel()

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name   string
			Key    string
			Id     string
			Secret string
		}

		tests := []Test{
			{"Simple KV", "password", "example.com", "username:password"},
		}

		for _, test := range tests {
			test := test
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				var key = []byte(test.Key)

				var original bytes.Buffer
				{
					var db = database.New()
					db.Set(test.Id, test.Secret)

					argon := crypto.DefaultArgon()
					defer argon.Release()
					err := db.Save(key, 1024, &argon, &original)
					if err != nil {
						t.Fatalf("expecting no errors, but received: %v", err)
					}
				}

				db, err := database.Open(key, bytes.NewReader(original.Bytes()))
				if err != nil {
					t.Fatalf("expecting no errors, but received: %v", err)
				}

				data, err := db.Get(test.Id)
				if err != nil {
					t.Fatalf("expecting no errors, but received: %v", err)
				}

				if test.Secret != data {
					t.Fatalf("expecting %s but received: %s", test.Secret, data)
				}

				err = db.Del(test.Id)
				if err != nil {
					t.Fatalf("expecting no errors, but received: %v", err)
				}

				_, err = db.Get(test.Id)
				if err == nil {
					t.Fatal("expecting error")
				}

				err = db.Del(test.Id)
				if err == nil {
					t.Fatal("expecting error")
				}
			})
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		t.Run("Invalid Secret JSON", func(t *testing.T) {
			_, err := database.Open(nil, strings.NewReader("["))
			if err == nil {
				t.Fatal("expecting error")
			}
		})

		t.Run("Wrong decryption key", func(t *testing.T) {
			key := []byte(t.Name())

			var original bytes.Buffer
			{
				argon := crypto.DefaultArgon()
				defer argon.Release()

				j := crypto.Job{
					Key:      key,
					Data:     []byte("{}"),
					Argon:    argon,
					SaltSize: 1024,
				}
				secret := j.Encrypt()
				err := json.NewEncoder(&original).Encode(secret)
				if err != nil {
					t.Fatalf("expecting no errors, but received: %v", err)
				}

			}

			_, err := database.Open([]byte("INVALID"), bytes.NewReader(original.Bytes()))
			if err == nil {
				t.Fatal("expecting error")
			}
		})

		t.Run("Invalid DB JSON", func(t *testing.T) {

			key := []byte(t.Name())

			var original bytes.Buffer
			{
				argon := crypto.DefaultArgon()
				defer argon.Release()

				j := crypto.Job{
					Key:      key,
					Data:     []byte("{"),
					Argon:    argon,
					SaltSize: 1024,
				}
				secret := j.Encrypt()
				err := json.NewEncoder(&original).Encode(secret)
				if err != nil {
					t.Fatalf("expecting no errors, but received: %v", err)
				}

			}

			_, err := database.Open(key, bytes.NewReader(original.Bytes()))
			if err == nil {
				t.Fatal("expecting error")
			}
		})
	})
}
