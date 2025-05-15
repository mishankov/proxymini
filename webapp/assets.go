package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:static
var assets embed.FS

func Assets() fs.FS {
	serverRoot, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}

	return serverRoot
}
