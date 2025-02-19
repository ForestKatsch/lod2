package cplane

import (
	"lod2/cplane/redeploy"
	"lod2/cplane/webhook"
	"log"
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

	r.Get("/redeploy", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Redeploying..."))
		log.Println("Redeploying...")
		redeploy.Redeploy()
	})

	return r
}
