package mount

import (
	"io"

	"github.com/RogueTeam/guardian/database"
)

const (
	Name = "guardian"
	Type = "guardian"
)

type Config struct {
	File     io.WriteSeeker
	Database *database.Database
}

func New(config Config) (f *FS, err error) {
	f = &FS{
		File:     config.File,
		Database: config.Database,
	}
	return
}
