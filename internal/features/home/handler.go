package home

import (
	"net/http"

	"github.com/nobuww/simpel-ktp/internal/store"
)

type Handler struct {
	Store *store.Store
}

func New(store *store.Store) *Handler {
	return &Handler{
		Store: store,
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	// ex: fetch users
	// users, err := h.Store.ListUsers(r.Context())

	Page().Render(r.Context(), w)
}

func (h *Handler) InfographicHandler(w http.ResponseWriter, r *http.Request) {
	Infographic().Render(r.Context(), w)
}

func (h *Handler) TrackerDemoHandler(w http.ResponseWriter, r *http.Request) {
	TrackerDemo().Render(r.Context(), w)
}
