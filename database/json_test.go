package database_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
)

func TestJson(t *testing.T) {
	t.Parallel()

	createTemp := func() (file *os.File, closeFunc func()) {
		file, err := os.CreateTemp("", "*")
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		closeFunc = func() {
			file.Close()
			os.RemoveAll(file.Name())
		}

		return file, closeFunc
	}
	t.Run("Secret management", func(t *testing.T) {
		t.Parallel()

		key := crypto.RandomData()
		defer key.Release()
		file, closeFunc := createTemp()
		defer file.Close()
		defer closeFunc()

		j, err := database.Open(file)
		if err != nil {
			t.Fatalf("Expecting no errors, received: %v", err)
		}

		fmt.Println(j)

	})
}
