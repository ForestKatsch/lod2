package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"lod2/cplane"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/hostrouter"
)

// TODO: environment variable
const port = 10800

func main() {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://beta.lod2.zip", "https://cplane.lod2.zip"}, // Use this to allow specific origin hosts
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Set up a static route
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	hr := hostrouter.New()

	hr.Map("cplane.lod2.zip", cplane.Router())

	r.Mount("/__cplane", cplane.Router())
	r.Mount("/", hr)

	fmt.Printf("Listening at localhost:%d\n", port)

	// Start the server
	http.ListenAndServe(fmt.Sprintf("localhost:%d", port), r)
}
