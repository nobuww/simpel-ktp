package errors

import (
	"net/http"
)

// Handler manages error page HTTP handlers
type Handler struct{}

// New creates a new error handler
func New() *Handler {
	return &Handler{}
}

// NotFoundHandler handles 404 errors
func (h *Handler) NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	NotFoundPage().Render(r.Context(), w)
}
