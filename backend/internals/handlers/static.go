package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

// SPAFileServer serves a built single-page app out of dir, falling back to
// dir/index.html for any path that isn't an actual file so client-side
// routing (react-router) keeps working on a hard refresh/deep link.
func SPAFileServer(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}
