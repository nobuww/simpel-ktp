package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"time"
)

type csrfContextKey string

const CSRFTokenKey csrfContextKey = "csrf_token"

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFFormName   = "csrf.Token"
)

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if cookie, err := r.Cookie(CSRFCookieName); err == nil {
			token = cookie.Value
		}

		if token == "" {
			token = generateRandomToken()
			setCSRFCookie(w, token)
		}

		ctx := context.WithValue(r.Context(), CSRFTokenKey, token)
		r = r.WithContext(ctx)

		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
			next.ServeHTTP(w, r)
			return
		}
		if !isValidOrigin(r) {
			http.Error(w, "Origin/Referer invalid", http.StatusForbidden)
			return
		}

		requestToken := r.Header.Get(CSRFHeaderName)
		if requestToken == "" {
			requestToken = r.FormValue(CSRFFormName)
		}

		if requestToken == "" || requestToken != token {
			http.Error(w, "CSRF token mismatch", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetCSRFToken(ctx context.Context) string {
	if val, ok := ctx.Value(CSRFTokenKey).(string); ok {
		return val
	}
	return ""
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func setCSRFCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,  // Not accessible by JS
		Secure:   false, // Set to true in production if HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}

func isValidOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin != "" {
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if u.Host == "localhost:8080" || u.Host == "127.0.0.1:8080" {
			return true
		}
		return u.Host == r.Host
	}

	referer := r.Header.Get("Referer")
	if referer != "" {
		u, err := url.Parse(referer)
		if err != nil {
			return false
		}
		if u.Host == "localhost:8080" || u.Host == "127.0.0.1:8080" {
			return true
		}
		return u.Host == r.Host
	}

	return false
}
