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
	"encoding/json"
	"fmt"
	"os"

	"github.com/RogueTeam/guardian/crypto"
)

type Database struct {
	File    *os.File                  `json:"-"`
	Secrets map[string]*crypto.Secret `json:"secrets"`
}

func (db *Database) Release() {
	db.File.Close()

	for _, secret := range db.Secrets {
		secret.Release()
	}
	clear(db.Secrets)
}

func (db *Database) SetSecret(id string, key, data *crypto.Data, argon *crypto.Argon) (err error) {
	secret, err := crypto.Encrypt(key, data, argon)
	if err != nil {
		err = fmt.Errorf("failed to encrypt: %w", err)
		return
	}

	if _, found := db.Secrets[id]; !found {
		db.Secrets[id] = new(crypto.Secret)
	}
	*db.Secrets[id] = secret
	return
}

func (db *Database) GetSecret(id string, key *crypto.Data) (data crypto.Data, err error) {
	secret, found := db.Secrets[id]
	if !found {
		return
	}
	data, err = crypto.Decrypt(key, secret)
	if err != nil {
		err = fmt.Errorf("failed to decrypt secret: %w", err)
		data.Release()
	}
	return
}

func (db *Database) DelSecret(id string) {
	secret, found := db.Secrets[id]
	if found {
		delete(db.Secrets, id)
		secret.Release()
	}
}

func (db *Database) Save() (err error) {
	_, err = db.File.Seek(0, 0)
	if err != nil {
		err = fmt.Errorf("failed to seek file: %w", err)
		return err
	}
	err = json.NewEncoder(db.File).Encode(db)
	if err != nil {
		err = fmt.Errorf("failed to encode file: %w", err)
	}
	return err
}

// Key is the master key for the database
// File is the final FS file to write all the changes
func Open(file *os.File) (db *Database, err error) {

	db = &Database{
		File: file,
	}

	stat, err := file.Stat()
	if err != nil {
		err = fmt.Errorf("failed to obtain stat from file: %w", err)
		return
	}

	if stat.Size() == 0 {
		db.Secrets = make(map[string]*crypto.Secret)
		return
	}

	err = json.NewDecoder(file).Decode(db)
	if err != nil {
		err = fmt.Errorf("failed to decode json: %w", err)
		db.Release()
		db = nil
	}

	return
}
