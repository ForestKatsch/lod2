package cplane

import (
	"lod2/cplane/webhook"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

var appStartTime time.Time

func init() {
	appStartTime = time.Now()
}

func Router() chi.Router {
	r := chi.NewRouter()

	r.Mount("/webhook/github", webhook.GithubWebhookRouter())

	// Define a basic route
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// Return number of seconds since the app started
		var uptime = time.Since(appStartTime).Seconds()
		w.Write([]byte("Uptime: " + (time.Duration(uptime) * time.Second).String()))
	})

	return r
}
