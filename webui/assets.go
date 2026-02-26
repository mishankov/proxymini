package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var assets embed.FS

func Assets() fs.FS {
	serverRoot, err := fs.Sub(assets, "build")
	if err != nil {
		panic(err)
	}

	return serverRoot
}
