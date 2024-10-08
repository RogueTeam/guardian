package mount

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/database"
)

type Dir struct {
	Inode    uint64
	File     IO
	Database *database.Database
}

var (
	_ fs.Node               = &Dir{}
	_ fs.HandleReadDirAller = &Dir{}
	_ fs.NodeStringLookuper = &Dir{}
	_ fs.NodeCreater        = &Dir{}
)

func (d *Dir) Attr(ctx context.Context, atr *fuse.Attr) (err error) {
	atr.Inode = d.Inode
	atr.Uid = uint32(os.Getuid())
	atr.Gid = uint32(os.Getgid())
	atr.Mode = os.ModeDir | 0o600
	return
}

func (d *Dir) ReadDirAll(ctx context.Context) (paths []fuse.Dirent, err error) {
	entries, err := d.Database.List()
	if err != nil {
		err = fmt.Errorf("failed to list secrets: %w", err)
		return
	}
	paths = make([]fuse.Dirent, len(entries))
	for index, entry := range entries {
		paths[index] = fuse.Dirent{
			Inode: uint64(index),
			Type:  fuse.DT_File,
			Name:  entry,
		}
	}
	return
}

func (d *Dir) Lookup(ctx context.Context, name string) (node fs.Node, err error) {
	found, err := d.Database.Lookup(name)
	if err != nil {
		err = fmt.Errorf("failed to lookup secret: %w", err)
		err = fmt.Errorf("%w: %w", err, syscall.EEXIST)
		return
	}

	if !found {
		err = syscall.ENOENT
		return
	}

	node = &File{
		Name:     name,
		Inode:    uint64(time.Now().UnixNano()),
		File:     d.File,
		Database: d.Database,
	}
	return
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (node fs.Node, h fs.Handle, err error) {
	log.Println("Creating")
	d.Database.Set(req.Name, "")
	node = &File{
		Name:     req.Name,
		Inode:    uint64(time.Now().UnixNano()),
		File:     d.File,
		Database: d.Database,
	}
	h = &Handle{
		Name:     req.Name,
		File:     d.File,
		Database: d.Database,
	}
	return
}
