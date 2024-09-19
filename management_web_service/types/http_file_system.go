package types

import (
	"io/fs"
	"net/http"
)

var _ fs.FS = &httpFs{}

type httpFs struct {
	fs http.FileSystem
}

func WrapHttpFsToOsFs(fs http.FileSystem) fs.FS {
	return &httpFs{
		fs: fs,
	}
}

func (h *httpFs) Open(name string) (fs.File, error) {
	return h.fs.Open(name)
}
