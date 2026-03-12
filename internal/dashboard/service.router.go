package dashboard

import "net/http"

type RouteRecord struct {
	Method  string
	Path    string
	Handler http.Handler
}

func (s *BoardSvc) Add(method string, path string, handler http.Handler) {
	s.routes = append(s.routes, RouteRecord{
		Method:  method,
		Path:    path,
		Handler: handler,
	})
}

func (s *BoardSvc) Get(path string, handler http.Handler) {
	s.Add(http.MethodGet, path, handler)
}

func (s *BoardSvc) Post(path string, hanler http.Handler) {
	s.Add(http.MethodPost, path, hanler)
}
