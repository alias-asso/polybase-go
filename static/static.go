package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed css/styles.css js/scripts.js svg
var content embed.FS

func FileSystem() http.FileSystem {
	fsys, err := fs.Sub(content, ".")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
