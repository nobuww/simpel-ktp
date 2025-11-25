package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nobuww/simpel-ktp/internal/handlers"
)

func main() {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir("./assets"))
	r.Handle("/assets/*", http.StripPrefix("/assets", fileServer))

	r.Get("/", handlers.HomeHandler)

	println("Server starting on :8080")
	http.ListenAndServe(":8080", r)
}
