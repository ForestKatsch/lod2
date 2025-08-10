package storage

import (
	"lod2/page"
	"lod2/storage"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func renderBrowseWithTemplate(w http.ResponseWriter, r *http.Request, path string, template string) {
	path, err := storage.VerifyPath(path)

	// the name of the file or directory as it appears in the UI.
	displayName := filepath.Base(path)

	// Check if we should serve raw file content
	raw := r.URL.Query().Get("raw") == "true"

	if raw {
		storage.ServeFile(w, r, path)
		return
	}

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	renderError := func(message string) {
		page.Render(w, r, "storage/index.html", map[string]interface{}{
			"Path":            path,
			"Name":            displayName,
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
		"Name":            displayName,
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
		metadata, err := storage.GetMetadata(path)
		if err != nil {
			renderError("failed to get file metadata")
			return
		}

		data["Size"] = metadata.Size
		data["LastModified"] = metadata.LastModified

		// Preview information.

		data["Type"] = mime.TypeByExtension(filepath.Ext(displayName))
	}

	page.Render(w, r, template, data)
}

func renderBrowsePath(w http.ResponseWriter, r *http.Request, path string) {
	renderBrowseWithTemplate(w, r, path, "storage/index.html")
}

func renderFileTable(w http.ResponseWriter, r *http.Request, path string) {
	renderBrowseWithTemplate(w, r, path, "storage/file-table.html")
}

func getBrowsePath(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	renderBrowsePath(w, r, path)
}
