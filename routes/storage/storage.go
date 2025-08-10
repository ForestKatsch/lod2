package storage

import (
	"lod2/auth"
	"lod2/middleware"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.AuthRoleRequiredMiddleware(auth.Storage))

	r.Get("/*", getBrowsePath)
	r.Post("/*", postUploadPath)
	r.Put("/*", putCreateDirectory)
	r.Delete("/*", deletePath)

	return r
}
