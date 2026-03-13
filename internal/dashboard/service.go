package dashboard

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/lopolopen/gap/options/dashboard"
	"github.com/lopolopen/gap/options/gap"
)

const (
	ContentType     = "Content-Type"
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html; charset=utf-8"
)

type BoardSvc struct {
	gapOpts *gap.Options
	opts    *dashboard.Options
	routes  []RouteRecord
}

func NewBoardSvc(gapOpts *gap.Options, opts *dashboard.Options) *BoardSvc {
	return &BoardSvc{
		gapOpts: gapOpts,
		opts:    opts,
	}
}

func (s *BoardSvc) HandleSPA() http.Handler {
	prefix := s.opts.NormalPrefix()

	sub, err := fs.Sub(StaticFiles, DistDir)
	var fileServer http.Handler
	if err == nil {
		fileServer = http.FileServer(http.FS(sub))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.URL.Path == prefix || r.URL.Path == prefix+"/" {
			http.Redirect(w, r, prefix+"/"+Index, http.StatusFound)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/"+Index) {
			data, err := fs.ReadFile(sub, Index)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set(ContentType, ContentTypeHTML)
			w.Write(data)
			return
		}

		http.StripPrefix(prefix, fileServer).ServeHTTP(w, r)
	})
}
