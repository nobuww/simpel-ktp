package common

import (
	"net/http"
)

// WriteError writes an error message to the response with the given status code.
// It uses a simple HTML format for the error message.
func WriteError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(`<div class="text-center py-8 text-red-500"><p>` + message + `</p></div>`))
}

// WriteNotFound writes a 404 Not Found message.
func WriteNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Data tidak ditemukan"
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`<div class="text-center py-8 text-slate-500"><p>` + message + `</p></div>`))
}
