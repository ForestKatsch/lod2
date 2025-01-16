package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

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

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	hr := hostrouter.New()

	hr.Map("", r)
	hr.Map("cplane.lod2.zip", cplane.Router())

	fmt.Printf("Listening at localhost:%d\n", port)

	// Start the server
	http.ListenAndServe(fmt.Sprintf(":%d", port), hr)
}
