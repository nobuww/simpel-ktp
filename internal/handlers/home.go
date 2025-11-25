package handlers

import (
	"net/http"

	"github.com/nobuww/simpel-ktp/views"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Layout("Simpel KTP", templates.Home())
	component.Render(r.Context(), w)
}
