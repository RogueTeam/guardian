package mount_test

import (
	"bytes"
	"os"
	"path"
	"testing"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/mount"
)

func Test_Read(t *testing.T) {
	t.Parallel()

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		const (
			secretId    = "secret"
			secretValue = "value"
			password    = "password"
		)

		var original bytes.Buffer
		{
			var db = database.New()
			db.Key = []byte(password)
			db.Argon = crypto.DefaultArgon()
			db.SaltSize = crypto.DefaultSaltSize
			db.Set(secretId, secretValue)

			argon := crypto.DefaultArgon()
			defer argon.Release()
			err := db.Save(&original)
			if err != nil {
				t.Fatalf("expecting no errors, but received: %v", err)
			}
		}

		var dbConfig = database.Config{
			Key:      []byte(password),
			Argon:    crypto.DefaultArgon(),
			SaltSize: crypto.DefaultSaltSize,
		}

		db, err := database.Open(dbConfig, &original)
		if err != nil {
			t.Fatalf("failed to open database: %s", err)
		}

		if err != nil {
			t.Fatalf("expecting no errors: %s", err)
		}

		var mountConfig mount.Config
		mountConfig.Database = db
		f, err := mount.New(mountConfig)
		if err != nil {
			t.Fatalf("failed to prepare filesystem: %s", err)
		}

		mountPoint, err := os.MkdirTemp("", "*")
		if err != nil {
			t.Fatalf("failed to create mount directory: %s", err)
		}
		defer func() {
			go os.RemoveAll(mountPoint)
		}()

		c, err := fuse.Mount(
			mountPoint,
			fuse.FSName(mount.Name),
			fuse.Subtype(mount.Type),
		)
		if err != nil {
			t.Fatalf("failed to mount: %s", err)
		}
		defer func() {
			go c.Close()
			go fuse.Unmount(mountPoint)
		}()

		go fs.Serve(c, f)

		target := path.Join(mountPoint, secretId)
		contents, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("failed to read secret: %s", err)
		}

		t.Log(string(contents))
		if string(contents) != secretValue {
			t.Fatal("Contents doesn't match")
		}
	})
}

func Test_Write(t *testing.T) {
	t.Parallel()

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		const (
			secretId    = "secret"
			secretValue = "value"
			password    = "password"
		)

		var original bytes.Buffer
		{
			var db = database.New()
			db.Key = []byte(password)
			db.Argon = crypto.DefaultArgon()
			db.SaltSize = crypto.DefaultSaltSize
			db.Set(secretId, secretValue)

			argon := crypto.DefaultArgon()
			defer argon.Release()
			err := db.Save(&original)
			if err != nil {
				t.Fatalf("expecting no errors, but received: %v", err)
			}
		}

		var dbConfig = database.Config{
			Key:      []byte(password),
			Argon:    crypto.DefaultArgon(),
			SaltSize: crypto.DefaultSaltSize,
		}

		db, err := database.Open(dbConfig, &original)
		if err != nil {
			t.Fatalf("failed to open database: %s", err)
		}

		if err != nil {
			t.Fatalf("expecting no errors: %s", err)
		}

		var mountConfig mount.Config
		mountConfig.File, err = os.CreateTemp("", "*")
		if err != nil {
			t.Fatal("failed to create fallback file")
		}
		mountConfig.Database = db
		f, err := mount.New(mountConfig)
		if err != nil {
			t.Fatalf("failed to prepare filesystem: %s", err)
		}

		mountPoint, err := os.MkdirTemp("", "*")
		if err != nil {
			t.Fatalf("failed to create mount directory: %s", err)
		}
		defer func() {
			go os.RemoveAll(mountPoint)
		}()

		c, err := fuse.Mount(
			mountPoint,
			fuse.FSName(mount.Name),
			fuse.Subtype(mount.Type),
		)
		if err != nil {
			t.Fatalf("failed to mount: %s", err)
		}
		defer func() {
			go c.Close()
			go fuse.Unmount(mountPoint)
		}()

		go fs.Serve(c, f)

		target := path.Join(mountPoint, secretId)
		contents, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("failed to read secret: %s", err)
		}

		t.Log(string(contents))
		if string(contents) != secretValue {
			t.Fatal("Contents doesn't match")
		}

		const newContent = "new value"
		err = os.WriteFile(target, []byte(newContent), 0o600)
		if err != nil {
			t.Fatalf("failed to write contents: %s", err)
		}

		value, err := db.Get(secretId)
		if err != nil {
			t.Fatalf("failed to get last value: %s", err)
		}
		if value != newContent {
			t.Fatalf("expecting %s but got %s", newContent, value)
		}
	})
}
