package mount

import (
	"context"
	"errors"
	"fmt"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/RogueTeam/guardian/database"
)

type Handle struct {
	Name string

	Buffer   []byte
	File     IO
	Database *database.Database
}

var (
	_ fs.Handle         = &Handle{}
	_ fs.HandleWriter   = &Handle{}
	_ fs.HandleReader   = &Handle{}
	_ fs.HandleReleaser = &Handle{}
)

func (h *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	log.Println("Writing")
	grow := req.Offset + int64(len(req.Data))
	if int64(len(h.Buffer)) < grow {
		newBuffer := make([]byte, grow)
		copy(newBuffer, h.Buffer)
		h.Buffer = newBuffer
	}
	resp.Size = copy(h.Buffer[req.Offset:], req.Data)
	h.Database.Set(h.Name, string(h.Buffer))
	return
}

func (h *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) (err error) {
	log.Println("Reading")
	if req.Offset > int64(len(h.Buffer)) {
		err = errors.New("index out of range")
		return
	}

	resp.Data = make([]byte, 0, req.Size)
	resp.Data = append(resp.Data, h.Buffer[req.Offset:min(int64(len(h.Buffer)), req.Offset+int64(req.Size))]...)

	return
}

func (h *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) (err error) {
	log.Println("Releasing")
	h.Database.Set(h.Name, string(h.Buffer))
	if h.File == nil {
		return
	}

	log.Println("Saving changes")
	h.File.Seek(0, 0)
	err = h.Database.Save(h.File)
	if err != nil {
		err = fmt.Errorf("failed to save changes in DB: %w", err)
		return
	}
	err = h.File.Sync()
	if err != nil {
		err = fmt.Errorf("failed to sync changes: %w", err)
		return
	}
	return
}
