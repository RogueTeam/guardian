package database

import (
	"fmt"
	"sort"
)


func (db *Database) Set(id string, data string) {
	db.Secrets[id] = data
}

func (db *Database) Lookup(id string) (found bool, err error) {
	_, found = db.Secrets[id]
	return
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

func (db *Database) List() (names []string, err error) {
	names = make([]string, 0, len(db.Secrets))
	for key := range db.Secrets {
		names = append(names, key)
	}
	sort.Strings(names)
	return
}
