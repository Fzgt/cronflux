package api

import (
	"embed"
	"io/fs"
	"net/http"
)

// webFS holds the embedded dashboard assets.
//
//go:embed web
var webFS embed.FS

// registerDashboard serves the single-page dashboard at the root and its static
// assets under /ui/.
func (s *Server) registerDashboard(mux *http.ServeMux) {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		s.log.Error("dashboard assets unavailable", "err", err)
		return
	}
	mux.Handle("GET /ui/", http.StripPrefix("/ui/", http.FileServer(http.FS(sub))))

	index, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		s.log.Error("dashboard index unavailable", "err", err)
		return
	}
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}
