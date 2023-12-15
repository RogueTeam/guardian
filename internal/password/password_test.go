package password_test

import (
	"bytes"
	"testing"

	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/internal/password"
)

func Test_New(t *testing.T) {
	if bytes.Equal(password.New(crypto.DataSize), password.New(crypto.DataSize)) {
		t.Fatalf("expected generated passwords be different")
	}
}
