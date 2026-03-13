package dashboard

import (
	"net/http"
	"strings"
)

//go:generate go tool shoot new -opt -short -type=Options

type Options struct {
	//shoot: def="/gap-dashboard"
	PathPrefix string

	Route func(method string, path string, hander http.Handler)
}

func (o *Options) NormalPrefix() string {
	prefix := "/" + strings.Trim(o.PathPrefix, "/")
	return prefix
}
