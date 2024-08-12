package mount

import (
	"context"
	"fmt"
	"io"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/database"
)

type File struct {
	Name  string
	Inode uint64

	File     io.WriteSeeker
	Database *database.Database
}

var (
	_ fs.Node            = &File{}
	_ fs.HandleReadAller = &File{}
	_ fs.NodeOpener      = &File{}
)

func (f *File) Attr(ctx context.Context, atr *fuse.Attr) (err error) {
	atr.Inode = f.Inode
	atr.Uid = uint32(os.Getuid())
	atr.Gid = uint32(os.Getgid())
	atr.Mode = 0o600
	sData, err := f.Database.Get(f.Name)
	if err != nil {
		err = fmt.Errorf("failed to read secret: %w", err)
		return
	}
	atr.Size = uint64(len(sData))
	return
}

func (f *File) ReadAll(ctx context.Context) (data []byte, err error) {
	sData, err := f.Database.Get(f.Name)
	if err != nil {
		err = fmt.Errorf("failed to read secret: %w", err)
		return
	}

	data = []byte(sData)
	return
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (h fs.Handle, err error) {
	sData, _ := f.Database.Get(f.Name)
	h = &Handle{
		Name:     f.Name,
		File:     f.File,
		Buffer:   []byte(sData),
		Database: *f.Database,
	}
	return
}
