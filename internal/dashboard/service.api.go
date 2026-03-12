package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/lopolopen/gap/internal/enum"
)

func (s *BoardSvc) HandleAPIs() []RouteRecord {
	s.Get("metas", s.GetMetas())
	return s.routes
}

type Meta struct {
	Type    enum.MetaType `json:"type"`
	Name    string        `json:"name"`
	Version string        `json:"version"`
}

var metas []*Meta

func AddMeta(typ enum.MetaType, name string, version string) {
	metas = append(metas, &Meta{
		Type:    typ,
		Name:    name,
		Version: version,
	})
}

func (s *BoardSvc) GetMetas() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentType, ContentTypeJSON)
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(metas); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
