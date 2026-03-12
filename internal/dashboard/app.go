package dashboard

import "embed"

//go:embed app/dist
var StaticFiles embed.FS

const (
	DistDir = "app/dist"
	Index   = "index.html"
)
