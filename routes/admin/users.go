package admin

import (
	"lod2/auth"
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	type getUserData struct {
		Users []auth.UserInfo
	}

	// users, err := auth.GetUsers()
	// if err != nil {
	// 	page.RenderError(w, r, err)
	// 	return
	// }
	page.Render(w, r, "admin/users/index.html", map[string]interface{}{
		"Users": nil,
	})
}

func userRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getUsers)

	return r
}
