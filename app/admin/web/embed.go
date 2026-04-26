package web

import "embed"

//go:embed all:dist
var assets embed.FS

func Assets() embed.FS {
	return assets
}
