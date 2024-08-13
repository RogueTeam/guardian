package mount

import (
	"io"

	"github.com/RogueTeam/guardian/database"
)

type IO interface {
	io.WriteSeeker
	Sync() (err error)
}

const (
	Name = "guardian"
	Type = "guardian"
)

type Config struct {
	File     IO
	Database *database.Database
}

func New(config Config) (f *FS, err error) {
	f = &FS{
		File:     config.File,
		Database: config.Database,
	}
	return
}
