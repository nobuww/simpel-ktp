package home

import (
	"net/http"

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
	Page().Render(r.Context(), w)
}

func (h *Handler) InfographicHandler(w http.ResponseWriter, r *http.Request) {
	Infographic().Render(r.Context(), w)
}

func (h *Handler) TrackerDemoHandler(w http.ResponseWriter, r *http.Request) {
	TrackerDemo().Render(r.Context(), w)
}
