package account

import (
	"lod2/auth"
	"lod2/page"
	"net/http"
)

func getInviteLinkFragment(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetCurrentUserInfo(r.Context())
	inviteId, err := auth.GetUserInviteId(userInfo.UserId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		page.RenderError(w, r, err)
		return
	}

	page.Render(w, r, "account/fragment-invite-link.html", map[string]interface{}{
		"InviteId":  inviteId,
		"InviteUrl": "https://lod2.zip/auth/invite/" + inviteId,
	})
}
