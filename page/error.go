package page

import (
	"fmt"
	"net/http"
)

func RenderError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	fmt.Println(err)
}
