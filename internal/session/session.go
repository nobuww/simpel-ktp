package session

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

const (
	SessionName = "simpel-ktp-session"

	KeyUserID   = "user_id"
	KeyUserType = "user_type"
	KeyUserName = "user_name"
	KeyUserRole = "user_role"

	UserTypeWarga   = "warga"
	UserTypePetugas = "petugas"

	RoleAdminKecamatan = "ADMIN_KECAMATAN"
	RoleAdminKelurahan = "ADMIN_KELURAHAN"
)

// Manager handles session operations
type Manager struct {
	store *sessions.CookieStore
}

// UserSession contains session data for an authenticated user
type UserSession struct {
	UserID   string
	UserType string
	UserName string
	UserRole string // Only for petugas: ADMIN_KECAMATAN or ADMIN_KELURAHAN
}

// New creates a new session manager with the given secret
// If secret is empty, it reads from SESSION_SECRET env var
func New(secret string) *Manager {
	if secret == "" {
		secret = os.Getenv("SESSION_SECRET")
	}
	if secret == "" {
		secret = "development-secret-change-in-production"
	}

	store := sessions.NewCookieStore([]byte(secret))
	// Determine if we should enforce secure cookies
	isSecure := os.Getenv("GO_ENV") == "production"
	if secureEnv := os.Getenv("HTTP_SECURE"); secureEnv != "" {
		isSecure = secureEnv == "true"
	}

	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 24 hours default
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	}

	return &Manager{store: store}
}

// SetWargaSession creates a session for warga (citizen) users
func (m *Manager) SetWargaSession(w http.ResponseWriter, r *http.Request, nik, namaLengkap string, remember bool) error {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return err
	}

	session.Values[KeyUserID] = nik
	session.Values[KeyUserType] = UserTypeWarga
	session.Values[KeyUserName] = namaLengkap
	session.Values[KeyUserRole] = ""

	if remember {
		session.Options.MaxAge = 86400 * 30 // 30 days
	} else {
		session.Options.MaxAge = 86400 // 24 hours
	}

	return session.Save(r, w)
}

// SetPetugasSession creates a session for petugas (officer) users
func (m *Manager) SetPetugasSession(w http.ResponseWriter, r *http.Request, petugasID, namaPetugas, role string, remember bool) error {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return err
	}

	session.Values[KeyUserID] = petugasID
	session.Values[KeyUserType] = UserTypePetugas
	session.Values[KeyUserName] = namaPetugas
	session.Values[KeyUserRole] = role

	if remember {
		session.Options.MaxAge = 86400 * 30 // 30 days
	} else {
		session.Options.MaxAge = 86400 // 24 hours
	}

	return session.Save(r, w)
}

// GetSession retrieves the current user session
// Returns nil if no valid session exists
func (m *Manager) GetSession(r *http.Request) *UserSession {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return nil
	}

	userID, ok := session.Values[KeyUserID].(string)
	if !ok || userID == "" {
		return nil
	}

	userType, _ := session.Values[KeyUserType].(string)
	userName, _ := session.Values[KeyUserName].(string)
	userRole, _ := session.Values[KeyUserRole].(string)

	return &UserSession{
		UserID:   userID,
		UserType: userType,
		UserName: userName,
		UserRole: userRole,
	}
}

// ClearSession removes the current session (logout)
func (m *Manager) ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := m.store.Get(r, SessionName)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	session.Values = make(map[interface{}]interface{})

	return session.Save(r, w)
}

// IsAuthenticated checks if the user has a valid session
func (m *Manager) IsAuthenticated(r *http.Request) bool {
	return m.GetSession(r) != nil
}

// IsWarga checks if the authenticated user is a warga
func (m *Manager) IsWarga(r *http.Request) bool {
	s := m.GetSession(r)
	return s != nil && s.UserType == UserTypeWarga
}

// IsPetugas checks if the authenticated user is a petugas
func (m *Manager) IsPetugas(r *http.Request) bool {
	s := m.GetSession(r)
	return s != nil && s.UserType == UserTypePetugas
}
