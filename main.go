package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"lod2/auth"
	"lod2/config"
	"lod2/cplane"
	"lod2/db"
	"lod2/middleware"
	"lod2/page"
	"lod2/routes"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/go-chi/hostrouter"
)

func main() {
	config.Init(true)
	db.Init()
	
	// Run global migrations before initializing other packages
	db.RunMigrations()
	
	auth.Init()

	// The primary router.
	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.StripSlashes)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://beta.lod2.zip", "https://cplane.lod2.zip"}, // Use this to allow specific origin hosts
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	r.Use(middleware.AuthRefreshMiddleware())

	// Set up a static route
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	hr := hostrouter.New()

	hr.Map("cplane.lod2.zip", cplane.Router())
	hr.Map("*", routes.Router())

	// Used for testing.
	r.Mount("/__cplane", cplane.Router())
	r.Mount("/", hr)

	r.NotFound(page.NotFound)

	// Start the server.
	address := fmt.Sprintf("%s:%d", config.Config.Http.Host, config.Config.Http.Port)
	log.Printf("lod2 started; listening at %s", address)

	err := http.ListenAndServe(address, r)

	if err != nil {
		panic(err)
	}
}
