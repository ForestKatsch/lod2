package storage

import (
	"errors"
	"io"
	"lod2/page"
	"lod2/storage"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func postUploadPath(w http.ResponseWriter, r *http.Request) {
	uploadDirectory := chi.URLParam(r, "*")

	r.ParseMultipartForm(80 << 20)

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		page.RenderStatus(w, r, http.StatusBadRequest, "could not read form")
		return
	}
	defer file.Close()

	uploadPath := filepath.Join(uploadDirectory, fileHeader.Filename)

	exists, err := storage.Exists(uploadPath)

	if err != nil || exists {
		page.RenderStatus(w, r, http.StatusConflict, "file already exists")
		return
	}

	dangerousFilesystemPath, err := storage.DangerousFilesystemPath(uploadPath)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	// copy to temp file
	tempFile, err := os.CreateTemp("", "upload-*.part")
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		page.RenderError(w, r, err)
	}

	tempFile.Write(fileBytes)

	// TODO: this isn't atomic, but it's good enough for now. it'd be hard to maliciously exploit.
	exists, err = storage.Exists(uploadPath)

	if err != nil || exists {
		page.RenderError(w, r, errors.New("file already exists"))
		return
	}

	log.Printf("renaming %s to %s", tempFile.Name(), dangerousFilesystemPath)
	os.Rename(tempFile.Name(), dangerousFilesystemPath)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
