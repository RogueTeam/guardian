package mount

import (
	"context"
	"errors"
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
	_ fs.HandleFlusher  = &Handle{}
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

	log.Println("- Setting")
	h.Database.Set(h.Name, string(h.Buffer))
	return nil
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

func (h *Handle) Flush(ctx context.Context, req *fuse.FlushRequest) (err error) {
	log.Println("Flushing")

	log.Println("- Setting")
	h.Database.Set(h.Name, string(h.Buffer))
	if h.File == nil {
		return
	}
	return nil
}

func (h *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) (err error) {
	log.Println("Releasing")

	log.Println("- Setting")
	h.Database.Set(h.Name, string(h.Buffer))
	if h.File == nil {
		return
	}

	return
}
