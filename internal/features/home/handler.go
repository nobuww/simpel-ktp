package home

import (
	"net/http"

	"github.com/nobuww/simpel-ktp/internal/middleware"
	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
)

// Handler manages home page HTTP handlers
type Handler struct {
	store *store.Store
}

// New creates a new home handler with the required dependencies
func New(s *store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	userSession := middleware.GetUserFromContext(r.Context())
	if userSession != nil {
		target := "/dashboard"
		if userSession.UserType == session.UserTypePetugas {
			target = "/admin"
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	Page().Render(r.Context(), w)
}

func (h *Handler) InfographicHandler(w http.ResponseWriter, r *http.Request) {
	Infographic().Render(r.Context(), w)
}

func (h *Handler) TrackerDemoHandler(w http.ResponseWriter, r *http.Request) {
	TrackerDemo().Render(r.Context(), w)
}
