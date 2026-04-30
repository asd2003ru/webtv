package webui

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var assets embed.FS

func Handler() http.Handler {
	dist, err := fs.Sub(assets, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	files := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			if err := serveIndex(dist, w); err != nil {
				http.Error(w, "frontend not built: run make web-build", http.StatusServiceUnavailable)
			}
			return
		}

		if _, err := dist.Open(r.URL.Path[1:]); err == nil {
			files.ServeHTTP(w, r)
			return
		}

		if err := serveIndex(dist, w); err != nil {
			http.Error(w, "frontend not built: run make web-build", http.StatusServiceUnavailable)
		}
	})
}

func serveIndex(dist fs.FS, w http.ResponseWriter) error {
	f, err := dist.Open("index.html")
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = io.Copy(w, f)
	return err
}
