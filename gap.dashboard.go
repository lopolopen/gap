package gap

import (
	"net/http"
	"path"

	"github.com/lopolopen/gap/internal/dashboard"
)

var boardSvc *dashboard.BoardSvc

func initDashboard(gapOpts *Options) {
	opts := gapOpts.Dashboard()
	if opts == nil {
		return
	}

	if boardSvc != nil {
		panic("dashboard must be initialized only once")
	}
	boardSvc = dashboard.NewBoardSvc(gapOpts, opts)

	if opts.Route == nil {
		panic("mount func is nil; you need to call dashboard.Mount when UseDashboard")
	}

	prefix := opts.NormalPrefix()

	for _, r := range boardSvc.HandleAPIs() {
		opts.Route(r.Method, r.Path, r.Handler)
	}
	//* must be put bottom
	opts.Route(http.MethodGet, path.Join(prefix, "*"), boardSvc.HandleSPA())
}
