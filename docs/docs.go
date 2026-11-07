package httpapi

import (
	"io/fs"
	"net/http"
)

func RegisterDocs(mux *http.ServeMux, fsys fs.FS) {
	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.FS(fsys))))
}
