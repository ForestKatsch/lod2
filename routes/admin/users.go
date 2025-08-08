package admin

import (
	"context"
	"lod2/auth"
	"lod2/page"
	"net/http"
	"strconv"

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

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	sessions, err := auth.AdminGetUserSessions(user.UserId)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	data := map[string]interface{}{
		"User":     user,
		"Sessions": sessions,
	}

	if user.InvitedByUserId != nil {
		invitedByUser, err := auth.AdminGetUserById(*user.InvitedByUserId)

		if err == nil {
			data["InvitedByUser"] = invitedByUser
		}
	}

	// Add this on to do string handling in the backend instead of the template.
	data["UserRoleStrings"] = auth.GetRoleStrings(user.Roles)

	page.Render(w, r, "admin/users/user/index.html", data)
}

func postUserEndAllSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	err := auth.AdminInvalidateAllSessions(user.UserId)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("0"))
}

func postUserResetInvites(w http.ResponseWriter, r *http.Request) {
	invitesLeftString := r.URL.Query().Get("to")

	if invitesLeftString == "" {
		invitesLeftString = "0"
	}

	invitesLeft, err := strconv.Atoi(invitesLeftString)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	user := r.Context().Value("user").(auth.UserSessionInfo)
	err = auth.AdminSetRemainingInvites(user.UserId, invitesLeft)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(strconv.Itoa(invitesLeft)))
}

func userRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getUsers)

	r.Route("/{userId}", func(r chi.Router) {
		r.Use(userCtx)
		r.Get("/", getUser)
		r.Post("/end-all-sessions", postUserEndAllSessions)
		r.Post("/reset-invites", postUserResetInvites)
	})

	return r
}
