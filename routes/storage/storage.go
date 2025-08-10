package storage

import (
	"lod2/page"
	"lod2/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := chi.URLParam(r, "*")
		path, err := storage.VerifyPath(path)

		if err != nil {
			page.RenderError(w, r, err)
			return
		}

		renderError := func(message string) {
			page.Render(w, r, "storage/index.html", map[string]interface{}{
				"Path":            path,
				"PathBreadcrumbs": storage.GetPathBreadcrumbs(path),
				"ErrorMessage":    message,
			})
		}

		exists, err := storage.Exists(path)
		if err != nil {
			page.RenderError(w, r, err)
			return
		}

		if !exists {
			renderError("this path does not exist")
			return
		}

		isDirectory, err := storage.IsDirectory(path)
		if err != nil {
			renderError("failed to check if this is a directory")
			return
		}

		data := map[string]interface{}{
			"Path":            path,
			"PathBreadcrumbs": storage.GetPathBreadcrumbs(path),
			"IsDirectory":     isDirectory,
		}

		if isDirectory {
			entries, err := storage.ListContents(path)
			if err != nil {
				renderError("failed to list contents")
				return
			}

			data["Entries"] = entries
		} else {
		}

		page.Render(w, r, "storage/index.html", data)
	})

	return r
}
