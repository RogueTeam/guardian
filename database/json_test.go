package database_test

import (
	"bytes"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
)

func TestJson(t *testing.T) {
	t.Parallel()

	t.Run("Secret management", func(t *testing.T) {
		t.Parallel()

		var key = []byte("password")
		const id = "github.com"
		const value = "username:password"

		var original bytes.Buffer
		{
			var db = database.New()
			db.Set(id, value)

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

		data, err := db.Get(id)
		if err != nil {
			t.Fatalf("expecting no errors, but received: %v", err)
		}

		if value != data {
			t.Fatalf("expecting %s but received: %s", value, data)
		}

		err = db.Del(id)
		if err != nil {
			t.Fatalf("expecting no errors, but received: %v", err)
		}

		_, err = db.Get(id)
		if err == nil {
			t.Fatal("expecting error")
		}
	})
}
