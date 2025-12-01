package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nobuww/simpel-ktp/internal/features/admin"
	"github.com/nobuww/simpel-ktp/internal/features/auth"
	"github.com/nobuww/simpel-ktp/internal/features/errors"
	"github.com/nobuww/simpel-ktp/internal/features/home"
	"github.com/nobuww/simpel-ktp/internal/features/user"
	"github.com/nobuww/simpel-ktp/internal/middleware"
	"github.com/nobuww/simpel-ktp/internal/session"
	"github.com/nobuww/simpel-ktp/internal/store"
)

func New(s *store.Store, sessionMgr *session.Manager) *chi.Mux {
	r := chi.NewRouter()

	authMiddleware := middleware.NewAuth(sessionMgr)
	r.Use(authMiddleware.InjectUser)

	// Custom 404 handler
	errorHandler := errors.New()
	r.NotFound(errorHandler.NotFoundHandler)

	staticServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", staticServer))

	// Home
	homeHandler := home.New(s)
	r.Get("/", homeHandler.HomeHandler)
	r.Get("/components/infographic", homeHandler.InfographicHandler)
	r.Get("/api/tracker-demo", homeHandler.TrackerDemoHandler)

	// Auth
	authHandler := auth.New(s, sessionMgr)

	// Public auth pages (redirect if already logged in)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RedirectIfAuthenticated)
		r.Get("/login", authHandler.LoginPageHandler)
		r.Get("/petugas/login", authHandler.LoginPetugasPageHandler)
		r.Get("/register", authHandler.RegisterPageHandler)
	})

	// Auth API endpoints
	r.Post("/auth/login", authHandler.HandleLogin)
	r.Post("/auth/login/petugas", authHandler.HandleLoginPetugas)
	r.Post("/auth/register", authHandler.HandleRegister)
	r.Post("/auth/logout", authHandler.HandleLogout)

	// Admin routes (protected)
	adminHandler := admin.New(s)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePetugas)
		r.Get("/admin", adminHandler.DashboardHandler)
		r.Get("/admin/permohonan", adminHandler.PermohonanHandler)
		r.Get("/admin/permohonan/{id}", adminHandler.PermohonanDetailHandler)
		r.Get("/admin/permohonan/{id}/status", adminHandler.PermohonanStatusFormHandler)
		r.Post("/admin/permohonan/update-status", adminHandler.UpdateStatusHandler)
		r.Get("/admin/jadwal", adminHandler.JadwalHandler)
	})

	// User routes (protected - warga only)
	userHandler := user.New(s)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequireWarga)
		r.Get("/dashboard", userHandler.DashboardHandler)
	})

	return r
}
