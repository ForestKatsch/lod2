package page

import (
	"log"
	"net/http"
)

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	// TODO: use templating
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	log.Print("rendering error: ", err)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Custom 404 Page", http.StatusNotFound)
}
