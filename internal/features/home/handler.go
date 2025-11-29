package home

import (
	"net/http"

	"github.com/nobuww/simpel-ktp/internal/store"
	"github.com/nobuww/simpel-ktp/ui/layouts"
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

	component := layouts.Base("Simpel KTP", Home())
	component.Render(r.Context(), w)
}
