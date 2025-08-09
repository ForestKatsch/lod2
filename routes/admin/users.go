package admin

import (
	"context"
	"lod2/auth"
	"lod2/middleware"
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
	users, err := auth.AdminGetAllUsers()
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	currentUser := auth.GetCurrentUserInfo(r.Context())
	var canCreateUsers bool
	if currentUser != nil {
		remaining, err := auth.AdminInvitesRemaining(currentUser.UserId)
		if err == nil {
			canCreateUsers = remaining > 0 || remaining == -1 // -1 means unlimited
		}
	}

	page.Render(w, r, "admin/users/index.html", map[string]interface{}{
		"Users":          users,
		"CanCreateUsers": canCreateUsers,
	})
}

func getUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	currentUser := auth.GetCurrentUserInfo(r.Context())
	sessions, err := auth.AdminGetUserSessions(user.UserId)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	data := map[string]interface{}{
		"User":          user,
		"Sessions":      sessions,
		"CanDeleteUser": currentUser != nil && user.UserId != currentUser.UserId,
	}

	if user.InvitedByUserId != nil {
		invitedByUser, err := auth.AdminGetUserById(*user.InvitedByUserId)

		if err == nil {
			data["InvitedByUser"] = invitedByUser
		}
	}

	data["Roles"] = user.Roles

	page.Render(w, r, "admin/users/user/index.html", data)
}

func deleteUserSessions(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	err := auth.AdminInvalidateAllSessions(user.UserId)

	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("0"))
}

func putUserResetInvites(w http.ResponseWriter, r *http.Request) {
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

func deleteUserDelete(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)
	currentUser := auth.GetCurrentUserInfo(r.Context())

	// Prevent users from deleting themselves
	if currentUser != nil && user.UserId == currentUser.UserId {
		http.Error(w, "Cannot delete yourself", http.StatusBadRequest)
		return
	}

	err := auth.AdminDeleteUser(user.UserId)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	w.Header().Set("Hx-Location", "/admin/users")
	w.WriteHeader(http.StatusOK)
}

func getCreateUser(w http.ResponseWriter, r *http.Request) {
	page.Render(w, r, "admin/users/create.html", map[string]interface{}{})
}

func postCreateUser(w http.ResponseWriter, r *http.Request) {
	currentUser := auth.GetCurrentUserInfo(r.Context())
	if currentUser == nil {
		page.Render401(w, r)
		return
	}

	r.ParseForm()

	username := r.Form.Get("username")
	password := r.Form.Get("password")
	confirmPassword := r.Form.Get("confirm_password")

	renderError := func(errorMsg string) {
		page.Render(w, r, "admin/users/create.html", map[string]interface{}{
			"Error":           errorMsg,
			"Username":        username,
			"Password":        password,
			"ConfirmPassword": confirmPassword,
		})
	}

	if username == "" {
		renderError("Username is required")
		return
	}

	if password == "" {
		renderError("Password is required")
		return
	}

	if password != confirmPassword {
		renderError("Passwords do not match")
		return
	}

	userId, err := auth.AdminInviteUser(currentUser.UserId, username, password)
	if err != nil {
		renderError(err.Error())
		return
	}

	http.Redirect(w, r, "/admin/users/"+userId, http.StatusSeeOther)
}

func putUserRoles(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(auth.UserSessionInfo)

	r.ParseForm()

	roles := []auth.Role{}

	for _, scope := range auth.AllAccessScopes {
		levelStr := r.Form.Get(strconv.Itoa(int(scope)))
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			page.RenderError(w, r, err)
			return
		}
		roles = append(roles, auth.Role{Scope: scope, Level: auth.AccessLevel(level)})
	}

	err := auth.AdminSetUserRoles(user.UserId, roles)
	if err != nil {
		page.RenderError(w, r, err)
		return
	}

	page.Render(w, r, "admin/users/user/fragment-roles-table.html", map[string]interface{}{
		"User":    user,
		"Roles":   roles,
		"Message": "User roles updated",
	})
}

func userRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.AuthRoleRequiredMiddleware(auth.UserManagement))

	r.Get("/", getUsers)
	r.Get("/create", getCreateUser)
	r.Post("/create", postCreateUser)

	r.Route("/{userId}", func(r chi.Router) {
		r.Use(userCtx)
		r.Get("/", getUser)
		r.Delete("/sessions", deleteUserSessions)
		r.Put("/invites", putUserResetInvites)
		r.Put("/roles", putUserRoles)
		r.Delete("/delete", deleteUserDelete)
	})

	return r
}
