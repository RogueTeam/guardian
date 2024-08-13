package mount_test

import (
	"bytes"
	"os"
	"testing"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/mount"
)

func Test_New(t *testing.T) {
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

		entries, err := os.ReadDir(mountPoint)
		if err != nil {
			t.Fatalf("failed to list entries in mount point: %s", err)
		}

		if len(entries) != 1 {
			t.Fatalf("failed to list entries, expecting 1 but got: %d", len(entries))
		}
		for _, entry := range entries {
			t.Log(entry.Name())
			t.Log(entry.Type())
			t.Log(entry.Info())
		}
	})
}
