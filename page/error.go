package page

import (
	"log"
	"net/http"
)

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	// TODO: use templating
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	log.Print("TODO rendering error: ", err)
}

func RenderStatus(w http.ResponseWriter, r *http.Request, status int, message string) {
	// TODO: use templating
	http.Error(w, message, status)
}

func Render401(w http.ResponseWriter, r *http.Request) {
	Render(w, r, "/_error/401.html", nil)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	Render(w, r, "/_error/404.html", nil)
}
