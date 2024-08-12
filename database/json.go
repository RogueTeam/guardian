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
	Secrets map[string]string `json:"secrets"`
}

func (db *Database) Set(id string, data string) {
	db.Secrets[id] = data
}

func (db *Database) Get(id string) (data string, err error) {
	data, found := db.Secrets[id]
	if !found {
		err = fmt.Errorf("no entry found with id: %s", id)
	}
	return
}

func (db *Database) Del(id string) (err error) {
	_, found := db.Secrets[id]
	if !found {
		err = fmt.Errorf("no entry found with id: %s", id)
		return
	}

	delete(db.Secrets, id)
	return
}

func (db *Database) Save(key []byte, saltSize int, argon *crypto.Argon, w io.Writer) (err error) {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(db)

	var job = crypto.Job{
		Key:      make([]byte, len(key)),
		Data:     buffer.Bytes(),
		Argon:    *argon,
		SaltSize: saltSize,
	}
	copy(job.Key, key)
	defer job.Release()

	secret := job.Encrypt()
	defer secret.Release()
	err = json.NewEncoder(w).Encode(secret)
	return
}

func New() (db *Database) {
	return &Database{
		Secrets: make(map[string]string),
	}
}

func Open(key []byte, r io.Reader) (db *Database, err error) {
	var secret crypto.Secret
	defer secret.Release()
	err = json.NewDecoder(r).Decode(&secret)
	if err != nil {
		err = fmt.Errorf("failed to decode secret: %w", err)
		return
	}

	var job = crypto.Job{
		Key: make([]byte, len(key)),
	}
	copy(job.Key, key)
	defer job.Release()
	err = job.Decrypt(&secret)
	if err != nil {
		err = fmt.Errorf("error during decryption: %w", err)
		return
	}

	db = New()
	err = json.Unmarshal(job.Data, db)
	if err != nil {
		err = fmt.Errorf("failed to decode JSON database: %w", err)
	}
	return
}
