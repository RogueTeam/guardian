// You may be asking yourself. Why not SQLite? PostgreSQL? XYZ?
// The reason is simple. Supply Chain Attacks? The more we depend on
// external packages, the greater the risks of those packages being exploited by someone else
// Even more when is something as critical as the data storage.
// Based on this statement, a built-in data storage is used. Using mainly encoding/json which is
// already maintained by the Go Team. That should be the only source of trust. Maybe? In an ideal
// world not even trust the compiler. But here we don't have other choice, we could write everything by our selfs (but is unreliable).
// If we could rank trust between authors, Go Team will rank always top. At least in the context of Go. And Google. inc
// is not secretly evil.
package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/RogueTeam/guardian/crypto"
)

type Database struct {
	Key      []byte
	SaltSize int
	Argon    crypto.Argon
	Secrets  map[string]string `json:"secrets"`
}

func New() (db *Database) {
	return &Database{
		Secrets: make(map[string]string),
	}
}

func (db *Database) Save(w io.Writer) (err error) {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(db)

	var job = crypto.Job{
		Key:      make([]byte, len(db.Key)),
		Data:     buffer.Bytes(),
		Argon:    db.Argon,
		SaltSize: db.SaltSize,
	}
	copy(job.Key, db.Key)
	defer job.Release()

	secret := job.Encrypt()
	defer secret.Release()
	err = json.NewEncoder(w).Encode(secret)
	return
}

type Config struct {
	Key      []byte
	Argon    crypto.Argon
	SaltSize int
}

func Open(config Config, r io.Reader) (db *Database, err error) {
	var secret crypto.Secret
	defer secret.Release()
	err = json.NewDecoder(r).Decode(&secret)
	if err != nil {
		err = fmt.Errorf("failed to decode secret: %w", err)
		return
	}

	var job = crypto.Job{
		Key: make([]byte, len(config.Key)),
	}
	if secret.Argon.Memory != 0 &&
		secret.Argon.Threads != 0 &&
		secret.Argon.Time != 0 {
		copy(job.Key, config.Key)
		defer job.Release()
		err = job.Decrypt(&secret)
		if err != nil {
			err = fmt.Errorf("error during decryption: %w", err)
			return
		}
	} else {
		job.Data = []byte("{}")
	}

	db = New()
	db.Key = config.Key
	db.Argon = config.Argon
	db.SaltSize = config.SaltSize
	err = json.Unmarshal(job.Data, db)
	if err != nil {
		err = fmt.Errorf("failed to decode JSON database: %w", err)
	}
	return
}
