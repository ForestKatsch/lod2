package admin

import (
	"context"
	"lod2/auth"
	"lod2/page"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func userCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := chi.URLParam(r, "userId")
		user, err := auth.AdminGetUserById(userId)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	type getUserData struct {
		Users []auth.UserInfo
	}

	users, err := auth.AdminGetAllUsers()
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	page.Render(w, r, "admin/users/index.html", map[string]interface{}{
		"Users": users,
	})
}

func postUserSignOut(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	err := auth.AdminInvalidateAllSessions(user.UserId)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func userRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getUsers)

	r.Route("/{userId}", func(r chi.Router) {
		r.Use(userCtx)
		r.Post("/sign-out", postUserSignOut)
	})

	return r
}
