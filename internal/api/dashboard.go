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
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("GET /ui/", http.StripPrefix("/ui/", fileServer))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})
}
