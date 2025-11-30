package auth

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

func (h *Handler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	LoginPage().Render(r.Context(), w)
}

func (h *Handler) LoginPetugasPageHandler(w http.ResponseWriter, r *http.Request) {
	LoginPetugasPage().Render(r.Context(), w)
}

func (h *Handler) RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	RegisterPage().Render(r.Context(), w)
}
