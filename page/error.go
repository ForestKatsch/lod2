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
