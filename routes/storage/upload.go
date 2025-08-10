package storage

import (
	"errors"
	"io"
	"lod2/page"
	"lod2/storage"
	"lod2/utils"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func postUploadPath(w http.ResponseWriter, r *http.Request) {
	uploadDirectory := chi.URLParam(r, "*")
	uploadDirectory, err := utils.UrlDecode(uploadDirectory)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

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

	err = storage.ImportFile(tempFile.Name(), uploadPath)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func putCreateDirectory(w http.ResponseWriter, r *http.Request) {
	createDirectory := chi.URLParam(r, "*")
	createDirectory, err := utils.UrlDecode(createDirectory)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	err = storage.CreateDirectory(createDirectory)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	// get all but the last component of uploadDirectory
	parentDirectory := filepath.Dir(createDirectory)

	renderFileTable(w, r, parentDirectory)
}

func deletePath(w http.ResponseWriter, r *http.Request) {
	deletePath := chi.URLParam(r, "*")
	deletePath, err := utils.UrlDecode(deletePath)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	err = storage.DeleteFile(deletePath)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	// get all but the last component of uploadDirectory
	parentDirectory := filepath.Dir(deletePath)

	renderFileTable(w, r, parentDirectory)
}
