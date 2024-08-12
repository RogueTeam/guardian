package mount

import (
	"io"
	"time"

	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/database"
)

type FS struct {
	File     io.WriteSeeker
	Database *database.Database
}

var _ fs.FS = &FS{}

func (f *FS) Root() (node fs.Node, err error) {
	node = &Dir{
		Inode:    uint64(uint64(time.Now().UnixNano())),
		File:     f.File,
		Database: f.Database,
	}
	return
}
