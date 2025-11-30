package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nobuww/simpel-ktp/internal/features/auth"
	"github.com/nobuww/simpel-ktp/internal/features/home"
	"github.com/nobuww/simpel-ktp/internal/store"
)

func New(s *store.Store) *chi.Mux {
	r := chi.NewRouter()

	staticServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", staticServer))

	// Home
	homeHandler := home.New(s)
	r.Get("/", homeHandler.HomeHandler)
	r.Get("/components/infographic", homeHandler.InfographicHandler)
	r.Get("/api/tracker-demo", homeHandler.TrackerDemoHandler)

	// Auth
	authHandler := auth.New(s)
	r.Get("/login", authHandler.LoginPageHandler)
	r.Get("/register", authHandler.RegisterPageHandler)
	// r.Post("/auth/login", authHandler.HandleLogin)    // future implementation
	// r.Post("/auth/register", authHandler.HandleRegister) // future implementation

	return r
}
