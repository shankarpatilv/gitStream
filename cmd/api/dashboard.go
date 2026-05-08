package main

import (
	"embed"
	"net/http"
)

//go:embed static/dashboard.html
var dashboardFiles embed.FS

func dashboardHandler() http.HandlerFunc {
	page, err := dashboardFiles.ReadFile("static/dashboard.html")
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(page)
	}
}
