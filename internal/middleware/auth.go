package middleware

import (
	"context"
	"net/http"

	"github.com/nobuww/simpel-ktp/internal/session"
)

type contextKey string

const UserContextKey contextKey = "user"

// Auth middleware dependencies
type Auth struct {
	Session *session.Manager
}

// NewAuth creates a new auth middleware instance
func NewAuth(sm *session.Manager) *Auth {
	return &Auth{Session: sm}
}

// InjectUser adds the current user session to the request context
// This middleware should be applied to all routes for template access
func (a *Auth) InjectUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSession := a.Session.GetSession(r)
		if userSession != nil {
			ctx := context.WithValue(r.Context(), UserContextKey, userSession)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

// RequireWarga ensures the user is authenticated as a warga (citizen)
// Redirects to /login if not authenticated or not a warga
func (a *Auth) RequireWarga(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Session.IsWarga(r) {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePetugas ensures the user is authenticated as a petugas (officer)
// Redirects to /petugas/login if not authenticated or not a petugas
func (a *Auth) RequirePetugas(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Session.IsPetugas(r) {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/petugas/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/petugas/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth ensures the user is authenticated (either warga or petugas)
// Redirects to /login if not authenticated
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Session.IsAuthenticated(r) {
			// Check if HTMX request
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RedirectIfAuthenticated redirects authenticated users away from login/register pages
func (a *Auth) RedirectIfAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userSession := a.Session.GetSession(r)
		if userSession != nil {
			redirectPath := "/"
			if userSession.UserType == session.UserTypePetugas {
				redirectPath = "/admin"
			}
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user session from context
func GetUserFromContext(ctx context.Context) *session.UserSession {
	user, ok := ctx.Value(UserContextKey).(*session.UserSession)
	if !ok {
		return nil
	}
	return user
}
