// Package dashboard provides the embedded web dashboard frontend.
package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed frontend/dist/*
var assets embed.FS

// Assets returns an fs.FS rooted at the frontend/dist directory.
func Assets() (fs.FS, error) {
	return fs.Sub(assets, "frontend/dist")
}
