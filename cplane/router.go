package cplane

import (
	"lod2/cplane/webhook"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/webhook/github", webhook.GithubWebhookRouter())

	// Define a basic route
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("200 OK"))
	})

	return r
}
