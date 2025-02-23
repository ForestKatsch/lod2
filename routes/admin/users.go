package admin

import (
	"lod2/internal/auth"
	"lod2/internal/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	type getUserData struct {
		Users []auth.UserInfo
	}
	users, err := auth.GetUsers()
	if err != nil {
		page.RenderError(w, r, err)
		return
	}
	page.Render(w, r, "admin/users/index.html", map[string]interface{}{
		"Users": users,
	})
}

func userRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getUsers)

	return r
}
