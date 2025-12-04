package common

import (
	"net/http"

	"github.com/nobuww/simpel-ktp/internal/middleware"
	"github.com/nobuww/simpel-ktp/internal/session"
)

// GetUserOrRedirect attempts to retrieve the user session from the request context.
// If the user is not found, it redirects to the specified login URL and returns (nil, false).
// If the user is found, it returns (user, true).
func GetUserOrRedirect(w http.ResponseWriter, r *http.Request, loginURL string) (*session.UserSession, bool) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, loginURL, http.StatusSeeOther)
		return nil, false
	}
	return user, true
}

// FormatRole converts the internal role string to a human-readable format.
func FormatRole(role string) string {
	switch role {
	case session.RoleAdminKecamatan:
		return "Admin Kecamatan"
	case session.RoleAdminKelurahan:
		return "Admin Kelurahan"
	default:
		return role
	}
}
